package main
import (
	"database/sql"
	_ "github.com/go-sql-driver/MySQL"
)


type DbMgr struct {
	db *sql.DB
}

var gDbMgr *DbMgr = nil

func InitGlobalDbMgr() {
	if gDbMgr  == nil {
		gDbMgr = &DbMgr {
			db: nil,
		}
	}
}

func GetDbMgr() *DbMgr {
	return gDbMgr
}

func (self *DbMgr)Init(connStr string) error {
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	self.db = db
	return nil
}

func (self *DbMgr)GetDbConnect() *sql.DB {
	return self.db
}