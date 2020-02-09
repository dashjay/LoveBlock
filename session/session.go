package session

import (
	"github.com/boltdb/bolt"
	"log"
)

const (
	DBUser = "user"
)

var db *bolt.DB

func init() {
	var err error
	db, err = bolt.Open("session.db", 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	var tables = []string{DBUser}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, l := range tables {
			_, err := tx.CreateBucketIfNotExists([]byte(l))
			if err != nil {
				panic(err)
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}
}

func GetDb() *bolt.DB {
	return db
}
