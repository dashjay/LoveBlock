package session

import (
	"github.com/boltdb/bolt"
	"log"
)

var (
	DBUser = []byte("user")
)

var db *bolt.DB

func init() {
	var err error
	db, err = bolt.Open("session.db", 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	var tables = [][]byte{DBUser}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, l := range tables {
			_, err := tx.CreateBucketIfNotExists(l)
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
