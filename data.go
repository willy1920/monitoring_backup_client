package main

import(
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Logs struct{

}

func (self *Config) SaveLog(kebun *string, timestamp *string, status *string) {
	sqlstmt := "INSERT INTO logs(kebun, last_update, status) VALUES(?,?,?)"
	stmt, err := self.db.Prepare(sqlstmt)
	checkErr(err)
	stmt.Exec(kebun, timestamp, status)
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