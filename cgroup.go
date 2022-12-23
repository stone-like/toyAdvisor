package main

import (
	"errors"
	"fmt"
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

	//今回はCpuStats.CpuUsage.TotalUsage,MemoryStats.MemoryUsage.TotalUsageのみ使用

	fmt.Println(stats)

	return &ContainerStats{
		Time:           time.Now(),
		CpuUsage:       stats.CpuStats.CpuUsage.TotalUsage,
		CpuSystemUsage: stats.CpuStats.CpuUsage.UsageInUsermode + stats.CpuStats.CpuUsage.UsageInKernelmode,
		MemoryUsage:    stats.MemoryStats.Usage.Usage - stats.MemoryStats.Stats["total_inactive_file"],
	}, nil
}
