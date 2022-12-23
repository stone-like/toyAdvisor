package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
)

var dockerStartingPoint = []string{
	"/sys/fs/cgroup/memory/docker",
	"/sys/fs/cgroup/cpu/docker",
}

type Advisor struct {
	containers  map[string]*Container
	SavePath    string
	MachineSpec *MachineSpec
}

func NewAdvisor(metricSavePath string) *Advisor {
	return &Advisor{
		containers: make(map[string]*Container),
		SavePath:   metricSavePath,
	}
}

// func (a *Advisor) GetDockerContainers() (map[string]struct{}, error) {
// 	containerNames := make(map[string]struct{})

// 	for _, path := range CgroupMounts {
// 		names, err := GetContainers(path, "/docker")
// 		if err != nil {
// 			return nil, err
// 		}

// 		for name := range names {
// 			containerNames[name] = struct{}{}
// 		}
// 	}

// 	return containerNames, nil

// }

type MachineSpec struct {
	NumCore        uint64
	MemoryCapacity uint64
}

func (a *Advisor) InitDockerContainers() error {

	containerNames, err := GetDockerContainers()
	if err != nil {
		return err
	}

	for name := range containerNames {
		container, err := CreateContainer(name, a.SavePath, a.MachineSpec)
		if err != nil {
			return err
		}
		a.containers[name] = container
	}

	return nil
}

var memoryCapacityRegexp = regexp.MustCompile(`MemTotal:\s*([0-9]+) kB`)

func GetMachineMemoryCapacity() (uint64, error) {
	out, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}

	memoryCapacity, err := parseCapacity(out, memoryCapacityRegexp)
	if err != nil {
		return 0, err
	}
	return memoryCapacity, err
}

func parseCapacity(b []byte, r *regexp.Regexp) (uint64, error) {
	matches := r.FindSubmatch(b)
	if len(matches) != 2 {
		return 0, fmt.Errorf("failed to match regexp in output: %q", string(b))
	}
	m, err := strconv.ParseUint(string(matches[1]), 10, 64)
	if err != nil {
		return 0, err
	}

	// Convert to bytes.
	return m * 1024, err
}

func (a *Advisor) GetMachineSpec() error {

	memcap, err := GetMachineMemoryCapacity()
	if err != nil {
		return err
	}

	a.MachineSpec = &MachineSpec{
		NumCore:        uint64(runtime.NumCPU()),
		MemoryCapacity: memcap,
	}

	return nil
}

func (a *Advisor) Start() {

	a.GetMachineSpec()

	//現在動いているdockerContainerを探してmetricを開始
	a.InitDockerContainers()

	////sys/.../docker以下を監視
	eventCh := make(chan string)

	for _, path := range dockerStartingPoint {
		watcher := NewWatcher(eventCh, path)
		watcher.Start()
	}

}

func (a *Advisor) Stop() {
	for _, container := range a.containers {
		container.Stop()
	}
}
