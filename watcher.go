package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type FsWatcher struct {
	watcher *fsnotify.Watcher
}

func NewFsWatcher() (*FsWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FsWatcher{watcher: watcher}, nil
}

func (f *FsWatcher) AddWatch(containerName string, dir string) error {
	path := filepath.Join(dir, containerName)
	fmt.Printf("watch start %s\n", path)
	return f.watcher.Add(path)
}

func (f *FsWatcher) Close() {
	f.watcher.Close()
}

//TODO chの型とファイル作成イベントをコンテナイベントに読み替える方法を学ぶ
type Watcher struct {
	stopCh      chan struct{}
	eventCh     chan Event
	cgroupPaths []string
	watcher     *FsWatcher
}

type ContainerEventType int

const (
	ContainerAdd ContainerEventType = iota
	ContainerDelete
)

type Event struct {
	EventType     ContainerEventType
	ContainerName string
}

func NewWatcher(eventCh chan Event, cgroupPaths []string) (*Watcher, error) {
	watcher, err := NewFsWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		cgroupPaths: cgroupPaths,
		stopCh:      make(chan struct{}),
		watcher:     watcher,
		eventCh:     eventCh,
	}, nil
}

func (w *Watcher) Start() error {

	for _, cgroupPath := range w.cgroupPaths {
		w.WatchDirectory(cgroupPath, "/")
	}

	go func() {
		for {
			select {
			case event := <-w.Event():
				fmt.Println("event Get")
				err := w.ProcessEvent(event)
				log.Println(err)
			case err := <-w.Error():
				log.Panicln(err)
			case <-w.stopCh:
				return
			}
		}
	}()

	return nil

}

func (w *Watcher) Stop() {
	close(w.stopCh)
	w.watcher.Close()
}

func (w *Watcher) ProcessEvent(event fsnotify.Event) error {

	var eventType ContainerEventType

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		eventType = ContainerAdd
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		eventType = ContainerDelete
	default:
		return nil
	}

	containerName := "/docker/" + filepath.Base(event.Name)
	containerEvent := Event{
		EventType:     eventType,
		ContainerName: containerName,
	}

	w.eventCh <- containerEvent

	return nil
}

//今回は/docker以下を監視するだけ、本家はdirを再帰で監視している
func (w *Watcher) WatchDirectory(dir string, containerName string) error {
	err := w.watcher.AddWatch(containerName, dir)
	if err != nil {
		return err
	}

	return nil
}

func (w *Watcher) Event() chan fsnotify.Event {
	return w.watcher.watcher.Events
}

func (w *Watcher) Error() chan error {
	return w.watcher.watcher.Errors
}
