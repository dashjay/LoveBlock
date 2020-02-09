package session

import (
	"github.com/boltdb/bolt"
	"log"
)

var db *bolt.DB

func init() {
	var err error
	db, err = bolt.Open("session.db", 0600, nil)
	if err != nil {
		log.Panic(err)
	}
}

func GetDb() *bolt.DB {
	return db
}
