package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAdvisor(t *testing.T) {

	path, err := os.Getwd()
	require.NoError(t, err)

	advisor := NewAdvisor(path)
	advisor.Start()

	time.Sleep(5 * time.Second)
	advisor.Stop()
}
