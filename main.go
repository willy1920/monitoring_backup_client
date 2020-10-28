package main

import (
	"log"
	"github.com/kardianos/service"
)

var logger service.Logger

type program struct{}

func (p *program) Start(s service.Service) error {
    // Start should not block. Do the actual work async.
    go p.run()
    return nil
}
func (p *program) run() {
	config := &Config{}
	config.InitSchedule()
}
func (p *program) Stop(s service.Service) error {
    // Stop should not block. Return with a few seconds.
    return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "Monitoring Backup",
		DisplayName: "Monitoring Backup",
		Description: "Monitoring Backup",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	checkErr(err)

	logger, err = s.Logger(nil)
	checkErr(err)
	
	err = s.Run()
	checkErr(err)
}

func checkErr(err error)  {
	if err != nil {
		log.Fatalf("%s", err)
	}
}