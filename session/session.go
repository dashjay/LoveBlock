package session

import (
	"errors"
	"github.com/boltdb/bolt"
	"log"
)

var db *bolt.DB

func init() {
	var err error
	db, err = bolt.Open("session.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}

func GetDb(name string) {
	err := db.Update(func(tx *bolt.Tx) error {

		//判断要创建的表是否存在
		b := tx.Bucket([]byte(name))
		if b == nil {
			return errors.New("db " + name + " not exists")
		}
		//一定要返回nil
		return nil
	})

	//更新数据库失败
	if err != nil {
		log.Fatal(err)
	}
}
