package main

import(
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type BackupLog struct{
	Kebun string
	Timestamp string
	Status string
}

func (self *Config) SaveLog(kebun *string, timestamp *string, status *string) {
	sqlstmt := "INSERT INTO logs(kebun, last_update, status) VALUES(?,?,?)"
	stmt, err := self.db.Prepare(sqlstmt)
	checkErr(err)
	stmt.Exec(kebun, timestamp, status)
}

func (self *Config) GetLogs() []BackupLog {
	rows, err := self.db.Query("SELECT kebun, last_update, status FROM logs")
	checkErr(err)
	defer rows.Close()

	var backupLog BackupLog
	var backupLogs []BackupLog
	for rows.Next(){
		rows.Scan(&backupLog.Kebun, &backupLog.Timestamp, &backupLog.Status)
		backupLogs = append(backupLogs, backupLog)
	}
	return backupLogs
}

func (self *Config) Init() {
	sqlstmt := `CREATE TABLE IF NOT EXISTS 'logs'(
		'kebun' CHAR(6),
		'last_update' CHAR(25),
		'status' CHAR(50),
		PRIMARY KEY('kebun', 'last_update')
	)`
	self.db, _ = sql.Open("sqlite3", "data.sqlite3")
	stmt, err := self.db.Prepare(sqlstmt)
	checkErr(err)
	stmt.Exec()
}