package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	path, _ := os.Getwd()

	advisor := NewAdvisor(path)
	err := advisor.Start()
	defer func() {
		advisor.Stop()
	}()

	if err != nil {
		log.Println(err)
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit

}
