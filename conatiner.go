package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

//cpuacct -> cpuStat,
//memory ->memoryStat
var CgroupMounts = map[string]string{
	"cpu":     "/sys/fs/cgroup/cpu",
	"cpuacct": "/sys/fs/cgroup/cpuacct",
	"memory":  "/sys/fs/cgroup/memory",
}

var dockerCgroupRegexp = regexp.MustCompile(`([a-z0-9]{64})`)

var collectInterval = 1 * time.Second

var memUnlimit uint64 = 9223372036854771712

type Status struct {
	PrevStatus *ContainerStats
	CurStatus  *ContainerStats
}

type ContainerSpec struct {
	MemoryLimit uint64
}

type Container struct {
	CgroupPaths   map[string]string
	Name          string
	CgroupManager *CgroupManager
	SavePath      string
	Stats         *Status
	StopCh        chan struct{}
	ContainerSpec *ContainerSpec
	MachineSpec   *MachineSpec
}

func isContainerName(name string) bool {
	// always ignore .mount cgroup even if associated with docker and delegate to systemd
	if strings.HasSuffix(name, ".mount") {
		return false
	}
	return dockerCgroupRegexp.MatchString(path.Base(name))
}

func MakeCgroupPath(containerName string, cgroupMounts map[string]string) map[string]string {
	m := make(map[string]string)

	for k, v := range cgroupMounts {
		m[k] = filepath.Join(v, containerName)
	}

	return m
}

func CreateContainer(containerName string, savePath string, machineSpec *MachineSpec) (*Container, error) {
	cgroupPaths := MakeCgroupPath(containerName, CgroupMounts)

	manager, err := NewCgroupManager(containerName, cgroupPaths)
	if err != nil {
		return nil, err
	}

	containerSpec := GetContainerSpec(cgroupPaths)

	container := &Container{
		CgroupPaths:   cgroupPaths,
		Name:          containerName,
		CgroupManager: manager,
		SavePath:      savePath,
		Stats:         &Status{},
		StopCh:        make(chan struct{}),
		MachineSpec:   machineSpec,
		ContainerSpec: containerSpec,
	}

	go container.Start()

	return container, nil

}

func GetContainerSpec(paths map[string]string) *ContainerSpec {
	spec := &ContainerSpec{}
	memoryPath, ok := paths["memory"]
	if !ok {
		return spec
	}
	if !FileExists(memoryPath) {
		return spec
	}

	spec.MemoryLimit = readUInt64(memoryPath, "memory.limit_in_bytes")

	return spec
}

func (c *Container) AddStats(stats *ContainerStats) {
	c.Stats.PrevStatus = c.Stats.CurStatus
	c.Stats.CurStatus = stats
}

func (c *Container) Stop() {
	close(c.StopCh)
}

func (c *Container) Start() {
	//metricの収集をスタート
	//tickerを使う
	go c.HouseKeeping()

}

func (c *Container) HouseKeeping() {

	//初回はすぐに発火
	timer := time.NewTimer(0 * time.Second)
	defer timer.Stop()

	for {

		if !c.DoHouseKeeping(timer.C) {
			//stopChがcloseしたとき
			return
		}

		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}

		timer.Reset(collectInterval)
	}
}

func (c *Container) DoHouseKeeping(timeCh <-chan time.Time) bool {
	select {
	case <-c.StopCh:
		return false
	case <-timeCh:
	}

	stats, _ := c.CgroupManager.GetStats()

	c.AddStats(stats)

	if c.Stats.PrevStatus != nil {
		//writeToFile
		c.WriteStatsToFile()
	}

	return true

}

//Time containerName
//memory:
// 0.3%
//cpu:
// 0.1%
//みたいにする
func (c *Container) WriteStatsToFile() error {
	if c.Stats.PrevStatus == nil {
		return nil
	}

	filename := fmt.Sprintf("metric%s.txt", filepath.Base(c.Name)[:12])

	f, err := os.OpenFile(filepath.Join(c.SavePath, filename), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return err
	}

	defer f.Close()

	status := c.CaluculateStats()
	fmt.Println(status)

	const format = "2006/01/02 15:04:05"
	str := fmt.Sprintf("%s %s\n", time.Now().Format(format), c.Name)
	_, err = f.WriteString(str)
	if err != nil {
		return err
	}

	str = fmt.Sprintf("memory:\n   %s\ncpu:\n   %s\n", status["memory"], status["cpu"])

	_, err = f.WriteString(str)
	if err != nil {
		return err
	}

	return nil
}

func (c *Container) CaluculateStats() map[string]string {
	// cpu := stats.CpuUsage
	// memory := stats.MemoryUsage
	prev := c.Stats.PrevStatus
	cur := c.Stats.CurStatus

	memoryLimit := c.ContainerSpec.MemoryLimit
	// numCPU := c.MachineSpec.NumCore
	cpuDelta := cur.CpuUsage - prev.CpuUsage
	interval := GetInterval(cur.Time, prev.Time)

	if memoryLimit == memUnlimit {
		memoryLimit = c.MachineSpec.MemoryCapacity
	}

	//ここのmemoryの計算はdocker statsと同じ
	memory := FormatFloat(((float64(cur.MemoryUsage)/float64(memoryLimit))*100), 2) + "%"

	//libcontainerではstatsを取ってくるときcpuStatsのsystem_cpu_usageを取ってきていないのでlibcontainerを使っていてはdocker statsと同じ値を出せなかったので、結局docker statsを呼ぶか、statsでやっているような/proc/statを合算しなくてはいけない
	//なので今回はcAdvisorで使われている手法で計算、ひとまず小数点二位のレベル(0.13)とか位までは一致するのでこれで行く

	//cAdvisorでは下記をさらにcore数で割った値を出していた
	cpu := FormatFloat(((float64(cpuDelta))/float64(interval)*100.0), 2) + "%"

	return map[string]string{
		"cpu":    cpu,
		"memory": memory,
	}
}

func GetDockerContainers() (map[string]struct{}, error) {
	containerNames := make(map[string]struct{})

	for _, path := range CgroupMounts {
		names, err := GetContainers(filepath.Join(path, "docker"), "/")
		if err != nil {
			return nil, err
		}

		for name := range names {
			if !isContainerName(name) {
				continue
			}

			name = filepath.Join("/docker", name)

			containerNames[name] = struct{}{}
		}
	}

	return containerNames, nil

}

//code from cAdvisor
func GetContainers(cgroupPath, containername string) (map[string]struct{}, error) {

	containers := make(map[string]struct{})
	err := ListDirectories(cgroupPath, containername, true, containers)

	if err != nil {
		return nil, err
	}

	return containers, nil

}
