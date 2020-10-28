package main

import (
	"log"
)

func main() {
	config := &Config{}
	config.InitSchedule()
}

func checkErr(err error)  {
	if err != nil {
		log.Fatalf("%s", err)
	}
}