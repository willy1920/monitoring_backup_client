package main

import (
	"log"

	"github.com/kardianos/service"
)

var logger service.Logger

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	config := Config{}
	config.InitSchedule()
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:					"Monitoring Backup",
		DisplayName:	"Monitoring Backup Service",
		Description:	"Monitoring Backup",
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