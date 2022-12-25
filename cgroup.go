package main

import (
	"errors"
	"time"

	"github.com/opencontainers/runc/libcontainer/cgroups"
	fs "github.com/opencontainers/runc/libcontainer/cgroups/fs"
	configs "github.com/opencontainers/runc/libcontainer/configs"
)

var (
	ErrInvalidMetrics = errors.New("InvalidMetricsError")
)

type CgroupManager struct {
	Manager cgroups.Manager
	eventCh chan Event
	// Handler
}

//pathsで欲しいStatsを指定しないとlibcontainerからStatsは得られないので注意
//例えばmemoryのみpathsに指定していた場合、cpuacctのStatはゼロ値のまま帰ってくる
func NewCgroupManager(containerName string, paths map[string]string) (*CgroupManager, error) {

	config := configs.Cgroup{
		Name:      containerName,
		Resources: &configs.Resources{},
	}
	manager, err := fs.NewManager(&config, paths)
	if err != nil {
		return nil, err
	}

	return &CgroupManager{
		Manager: manager,
	}, nil
}

type ContainerStats struct {
	Time           time.Time
	CpuUsage       uint64
	CpuSystemUsage uint64
	MemoryUsage    uint64
}

func (c *CgroupManager) GetStats() (*ContainerStats, error) {

	stats, err := c.Manager.GetStats()

	if err != nil {
		return nil, err
	}

	return &ContainerStats{
		Time:           time.Now(),
		CpuUsage:       stats.CpuStats.CpuUsage.TotalUsage,
		CpuSystemUsage: stats.CpuStats.CpuUsage.UsageInUsermode + stats.CpuStats.CpuUsage.UsageInKernelmode,
		MemoryUsage:    stats.MemoryStats.Usage.Usage - stats.MemoryStats.Stats["total_inactive_file"],
	}, nil
}

func (c *CgroupManager) WatchContainers() error {
	// go func() {
	// 	for {
	// 		select {
	// 		case event := <-c.eventCh:
	// 			switch {
	// 			case event.EventType == ContainerAdd:
	// 				err := CreateContainer(event.Name, event.WatchSource)
	// 				if err != nil {
	// 					log.Println(err)
	// 				}

	// 			case event.EventType == ContainerDelete:
	// 				// err := m.destroyContainer(event.Name)
	// 				// if err != nil {
	// 				// 	log.Println(err)
	// 				// }
	// 			}

	// 			// case <-quit:
	// 			// 	var errs partialFailure

	// 			// 	// Stop processing events if asked to quit.
	// 			// 	for i, watcher := range m.containerWatchers {
	// 			// 		err := watcher.Stop()
	// 			// 		if err != nil {
	// 			// 			errs.append(fmt.Sprintf("watcher %d", i), "Stop", err)
	// 			// 		}
	// 			// 	}

	// 			// 	if len(errs) > 0 {
	// 			// 		quit <- errs
	// 			// 	} else {
	// 			// 		quit <- nil
	// 			// 		klog.Infof("Exiting thread watching subcontainers")
	// 			// 		return
	// 			// 	}
	// 			// }
	// 		}
	// 	}
	// }()

	return nil
}
