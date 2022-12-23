package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	path, _ := os.Getwd()

	advisor := NewAdvisor(path)
	advisor.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit

	advisor.Stop()

}

// func GetContainer() (map[string]struct{}, error) {
// 	cPath := "/sys/fs/cgroup/cpu/"
// 	name := "/"
// 	containers := make(map[string]struct{})
// 	err := ListDirectories(cPath, name, true, containers)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return containers, nil

// }

// func ListDirectories(dirpath string, parent string, recursive bool, output map[string]struct{}) error {
// 	buf := make([]byte, godirwalk.MinimumScratchBufferSize)
// 	return listDirectories(dirpath, parent, recursive, output, buf)
// }

// func listDirectories(dirpath string, parent string, recursive bool, output map[string]struct{}, buf []byte) error {
// 	dirents, err := godirwalk.ReadDirents(dirpath, buf)
// 	if err != nil {
// 		// Ignore if this hierarchy does not exist.
// 		if os.IsNotExist(errors.Cause(err)) {
// 			err = nil
// 		}
// 		return err
// 	}
// 	for _, dirent := range dirents {
// 		// We only grab directories.
// 		if !dirent.IsDir() {
// 			continue
// 		}
// 		dirname := dirent.Name()

// 		name := path.Join(parent, dirname)
// 		output[name] = struct{}{}

// 		// List subcontainers if asked to.
// 		if recursive {
// 			err := listDirectories(path.Join(dirpath, dirname), name, true, output, buf)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// var dockerCgroupRegexp = regexp.MustCompile(`([a-z0-9]{64})`)

// func isContainerName(name string) bool {
// 	// always ignore .mount cgroup even if associated with docker and delegate to systemd
// 	if strings.HasSuffix(name, ".mount") {
// 		return false
// 	}
// 	return dockerCgroupRegexp.MatchString(path.Base(name))
// }

// func ContainerNameToDockerId(name string) string {
// 	id := path.Base(name)

// 	if matches := dockerCgroupRegexp.FindStringSubmatch(id); matches != nil {
// 		return matches[1]
// 	}

// 	return id
// }

// func CanHandleAndAccept(name string) (string, error) {
// 	// if the container is not associated with docker, we can't handle it or accept it.
// 	if !isContainerName(name) {
// 		return "", nil
// 	}

// 	// Check if the container is known to docker and it is active.
// 	id := ContainerNameToDockerId(name)

// 	return id, nil
// }
