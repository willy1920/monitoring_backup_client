package main

import (
	"log"
	"net/http"

	"github.com/kardianos/service"
	"github.com/mattn/go-ieproxy"
)

var logger service.Logger

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}
func (p *program) run() {
	// Do work here
	config := &Config{}
	config.InitSchedule()
}
func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	config := &Config{}
	config.StopAll()
	return nil
}

func main() {
	http.DefaultTransport.(*http.Transport).Proxy = ieproxy.GetProxyFunc()
	svcConfig := &service.Config{
		Name:        "Monitoring Backup",
		DisplayName: "Monitoring Backup",
		Description: "Monitoring Backup",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}

func checkErr(err error)  {
	if err != nil {
		log.Fatalf("%s", err)
	}
}