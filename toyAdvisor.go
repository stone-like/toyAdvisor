package main

import (
	"fmt"
	"log"
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
	watchers    []*Watcher
	stopCh      chan struct{}
	SavePath    string
	MachineSpec *MachineSpec
}

func NewAdvisor(metricSavePath string) *Advisor {
	return &Advisor{
		containers: make(map[string]*Container),
		SavePath:   metricSavePath,
		stopCh:     make(chan struct{}),
	}
}

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
		container, err := a.CreateContainer(name)
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

func (a *Advisor) Start() error {

	a.GetMachineSpec()

	//現在動いているdockerContainerを探してmetric収集を開始
	a.InitDockerContainers()

	////sys/.../docker以下を監視
	eventCh := make(chan Event)

	watcher, err := NewWatcher(eventCh, dockerStartingPoint)

	if err != nil {
		return err
	}

	watcher.Start()

	a.ProcessContainerEvent(eventCh)

	return nil

}

func (a *Advisor) ProcessContainerEvent(eventCh chan Event) {
	go func() {
		for {
			select {
			case event := <-eventCh:
				switch event.EventType {
				case ContainerAdd:
					_, exist := a.containers[event.ContainerName]
					if exist {
						continue
					}
					container, err := a.CreateContainer(event.ContainerName)
					if err != nil {
						log.Println(err)
						continue
					}
					a.containers[event.ContainerName] = container
				case ContainerDelete:
					_, exist := a.containers[event.ContainerName]
					if !exist {
						continue
					}
					err := a.DeleteContainer(event.ContainerName)
					if err != nil {
						log.Println(err)
					}
				}
			case <-a.stopCh:
				return
			}
		}
	}()
}

func (a *Advisor) CreateContainer(containerName string) (*Container, error) {
	fmt.Printf("createContainer %s\n", containerName)
	return CreateContainer(containerName, a.SavePath, a.MachineSpec)
}

func (a *Advisor) DeleteContainer(containerName string) error {
	fmt.Printf("deleteContainer %s\n", containerName)
	target := a.containers[containerName]

	target.Stop()

	delete(a.containers, containerName)
	return nil
}

func (a *Advisor) Stop() {
	for _, container := range a.containers {
		container.Stop()
	}

	for _, watcher := range a.watchers {
		watcher.Stop()
	}

	close(a.stopCh)
}
