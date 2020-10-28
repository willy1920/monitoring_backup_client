package main

import(
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
	WaitReconnect int `json:"waitReconnect"`
	db *sql.DB
	WatcherChan chan bool
	Schedule chan bool
	ScheduleRunning bool
}

func (self *Config) InitSchedule() {
	log.Println("Start watcher")
	self.readConfig()
	self.Init()
	self.Schedule = make(chan bool)
	self.ScheduleRunning = false

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	self.WatcherChan = make(chan bool)
	go func(){
		for{
			select{
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				switch event.Op {
				case fsnotify.Create:
					self.CheckName(strings.ToLower(filepath.Base(event.Name)))
				}
			case err, ok := <-watcher.Errors:
				if !ok{
					return
				}
				log.Println("Error: ", err)
			case <- self.WatcherChan:
				log.Println("Stop watcher")
				return
			}
		}
	}()

	err = watcher.Add(self.Path)
	if err != nil{
		log.Fatal(err)
	}
	<-self.WatcherChan
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
						if self.ScheduleRunning {
							self.Schedule <- true
							self.ScheduleCheckServer()
						} else{
							self.ScheduleCheckServer()
						}
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

func (self *Config) ScheduleCheckServer() chan bool {
	self.Schedule = make(chan bool)
	self.ScheduleRunning = true

	ticker := time.NewTicker(time.Duration(self.WaitReconnect) * time.Second)
	stop := make(chan bool)
	go func(){
		for{
			select {
			case <- ticker.C:
				log.Println("Start check server")
				self.ScheduleDeleteLogs()
			case <- self.Schedule:
				log.Println("Stop check server")
				ticker.Stop()
				return
			}
		}
	}()
	return stop
}

func (self *Config) ScheduleDeleteLogs() {
	var backupLogs = self.GetLogs()
	status := true

	for _, v := range backupLogs{
		ok := self.Created(v.Kebun, v.Timestamp, v.Status)
		if ok != nil {
			status = false
			break;
		} else {
			self.DeleteLog(&v.Kebun, &v.Timestamp)
		}
	}
	if status {
		self.ScheduleRunning = false
		self.Schedule <- true
	}
}