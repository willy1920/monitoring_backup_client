package main

import (
	"log"
	"os"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	"github.com/willy1920/monitoring_backup_proto_go"
	"github.com/fsnotify/fsnotify"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Config struct{
	Kebun []string `json:"kebun"`
	Path string `json:"path"`
	Server string `json:"server"`
}

func main() {
	config := Config{}
	config.readConfig()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func(){
		for{
			select{
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				switch event.Op {
				case fsnotify.Create:
					config.CheckName(strings.ToLower(filepath.Base(event.Name)))
				}
				// log.Println("Event: ", event)
				// if event.Op&fsnotify.Write == fsnotify.Write {
				// 	log.Println("Modified file: ", event.Name)
				// }
			case err, ok := <-watcher.Errors:
				if !ok{
					return
				}
				log.Println("Error: ", err)
			}
		}
	}()

	err = watcher.Add(config.Path)
	if err != nil{
		log.Fatal(err)
	}
	<-done
}

func (self *Config) readConfig() {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal([]byte(byteValue), &self)
}

func (self *Config) CheckName(name string) {
	for _, val := range self.Kebun {
		if strings.Contains(name, strings.ToLower(val)) {
			ext := filepath.Ext(name)
			filename := string(name[0:len(name)-len(ext)])
			filedatestring := string(filename[len(filename)-8:])
			layout := "01022006"
			layoutSend := "2006-01-02 15:04:05"
			
			switch ext {
			case ".dmp":
				if filedatestring == time.Now().Format(layout) {
					self.Created(val, time.Now().Format(layoutSend), "Exporting Dump File")
				}
			case ".rar":
				log.Println("rar")
			}
			break;
		}
	}
}

func (self *Config) Created(kebun string, timestamp string, status string) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(self.Server, grpc.WithInsecure())
	if err != nil{
		log.Fatalf("Did not connect: %s", err)
	}
	defer conn.Close()

	c := monitoring_backup.NewMonitoringBackupClient(conn)
	
	response, err := c.SendNotif(context.Background(), &monitoring_backup.CreatedNotify{Kebun: kebun, Timestamp: timestamp, Status: status})
	if err != nil {
		log.Fatalf("Error when calling SendNotif: %s", err)
	}
	log.Printf("Response from server: %s", response.Kebun)
}