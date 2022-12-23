package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {

	// subSystems, err := GetAllCgroupSubsystems()
	// require.NoError(t, err)

	// Generate the equivalent cgroup manager for this container.
	cgroupManager, err := NewCgroupManager("/docker/7704a07571226d8050ed2439aea69d6fc1a09a6cafa95dabc4ff13539607f58b", CgroupMounts)
	require.NoError(t, err)

	cgroupManager.GetStats()

}
