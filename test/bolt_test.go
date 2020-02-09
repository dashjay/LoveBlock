package test

import (
	"github.com/boltdb/bolt"
	"main/session"
	"testing"
)

func TestBolt(t *testing.T) {
	db := session.GetDb()

	err := db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("test"))
		if b != nil {
			err := b.Put([]byte("test"), []byte("test"))
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		t.Error(err)
	}
}
