package main

import (
	"log"
	"sync"

	"github.com/judwhite/go-svc/svc"
)

// program implements svc.Service
type program struct {
	wg   sync.WaitGroup
	quit chan struct{}
	config Config
}

func main() {
	prg := &program{}

	// Call svc.Run to start your program/service.
	if err := svc.Run(prg); err != nil {
		log.Fatal(err)
	}
}

func (p *program) Init(env svc.Environment) error {
	log.Printf("is win service? %v\n", env.IsWindowsService())
	return nil
}

func (p *program) Start() error {
	// The Start method must not block, or Windows may assume your service failed
	// to start. Launch a Goroutine here to do something interesting/blocking.

	p.quit = make(chan struct{})

	p.wg.Add(1)
	go func() {
		log.Println("Starting...")

		p.config.InitSchedule()

		<-p.quit
		log.Println("Quit signal received...")
		p.wg.Done()
	}()

	return nil
}

func (p *program) Stop() error {
	log.Println("Stopping...")
	
	close(p.config.Schedule)
	close(p.config.WatcherChan)

	close(p.quit)
	p.wg.Wait()
	log.Println("Stopped.")
	return nil
}

func checkErr(err error)  {
	if err != nil {
		log.Fatalf("%s", err)
	}
}