package main

import (
	"log"
	"os"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"
	"database/sql"

	"github.com/willy1920/monitoring_backup_proto_go"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Config struct{
	Kebun []string `json:"kebun"`
	Path string `json:"path"`
	Server string `json:"server"`
	db *sql.DB
}

func main() {
	config := Config{}
	config.readConfig()
	config.Init()

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
					ok := self.Created(val, time.Now().Format(layoutSend), "Exporting Dump File")
					if ok != nil {
						self.ScheduleCheckServer()
					}
				}
			case ".rar":
				if filedatestring == time.Now().Format(layout) {
					self.Created(val, time.Now().Format(layoutSend), "Success")
				}
			}
			break;
		}
	}
}

func (self *Config) Created(kebun string, timestamp string, status string) error {
	log.Println("Dialing server...")
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(self.Server, grpc.WithInsecure())
	if err != nil{
		self.SaveLog(&kebun, &timestamp, &status)
		log.Println("Did not connect: %s", err)
		return err
	}
	defer conn.Close()

	log.Println("Making gRPC request...")
	c := monitoring_backup.NewMonitoringBackupClient(conn)
	
	log.Println("Send request")
	response, err := c.SendNotif(context.Background(), &monitoring_backup.CreatedNotify{Kebun: kebun, Timestamp: timestamp, Status: status})
	if err != nil {
		self.SaveLog(&kebun, &timestamp, &status)
		log.Println("Error when calling SendNotif: %s", err)
		return err
	} else{
		log.Printf("Response from server: %s", response.Kebun)
	}
	return nil
}

func (self *Config) ScheduleCheckServer() {
	var backupLogs = self.GetLogs()
	log.Println(backupLogs)

	ticker := time.NewTicker(2 * time.Second)
	quit := make(chan struct{})
	go func(){
		for{
			select {
			case <- ticker.C:
				log.Println("a")
				for _, v := range backupLogs{
					ok := self.Created(v.Kebun, v.Timestamp, v.Status)
					if ok != nil {
						break;
					}
				}
			case <- quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func checkErr(err error)  {
	if err != nil {
		log.Fatalf("%s", err)
	}
}