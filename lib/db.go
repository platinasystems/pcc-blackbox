package pcc

import (
	"fmt"
	"github.com/jinzhu/gorm"
	log "github.com/platinasystems/go-common/logs"
	"strings"
)

type DBHandler struct {
}

type DBConfiguration struct {
	Address string `"json:address"`
	Port    int    `"json:port"`
	Name    string `"json:name"`
	User    string `"json:user"`
	Pwd     string `"json:pwd"`
}

var dbConfig *DBConfiguration
var DB *gorm.DB

func InitDB(c *DBConfiguration) {
	if c == nil {
		return
	}

	dbConfig = c

	if len(strings.TrimSpace(dbConfig.Address)) > 0 {
		log.AuctaLogger.Infof("init database handler", dbConfig.Address, dbConfig.Port, dbConfig.Name)

		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			dbConfig.Address,
			dbConfig.Port,
			dbConfig.User,
			dbConfig.Pwd,
			dbConfig.Name)

		db, err := gorm.Open("postgres", psqlInfo)
		if err == nil {
			db.DB()
			db.DB().Ping()
			db.DB().SetMaxIdleConns(10)
			db.DB().SetMaxOpenConns(100)
			DB = db
		} else {
			log.AuctaLogger.Errorf("unable to start database. Fatal Exception: [%+v]", err.Error())
			panic(err)
		}
	}
}

func (dbh *DBHandler) getDB() *gorm.DB {
	return DB
}
