package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"

	"github.com/squeed/geezo-bot/pkg/config"
)

var conn *Conn

type Conn struct {
	db *gorm.DB
}

func GetConn(conf *config.Config) (*Conn, error) {
	// assumes config doesn't change.
	if conn != nil {
		return conn, nil
	}

	db, err := gorm.Open("sqlite3", conf.Main.DbFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open DB")
	}

	err = db.AutoMigrate(Image{}).Error
	if err != nil {
		return nil, errors.Wrap(err, "unable to migrate Image table")
	}

	conn = &Conn{db: db}

	return conn, nil
}

func (conn *Conn) Close() error {
	return conn.db.Close()
}
