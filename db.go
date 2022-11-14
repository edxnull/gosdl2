package main

import (
	"bytes"
	"fmt"
	"os"

	bolt "go.etcd.io/bbolt"
)

var FILE_MODE_RW os.FileMode = 0600

func DBOpen() *bolt.DB {
	wdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	db, err := bolt.Open(wdir+"/my.db", FILE_MODE_RW, nil)
	if err != nil {
		panic(err)
	}
	return db
}

func DBInit(db *bolt.DB, mk DBEntry) error {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("TestWords"))
		if err != nil {
			return fmt.Errorf("Failed to create bucket: %v", err)
		}
		for k, _ := range mk {
			if bucket.Get([]byte(k)) == nil { // don't override key/val if exists
				bt := [][]byte{
					[]byte(mk[k].Value),
					[]byte(mk[k].Tags[0]),
					[]byte(mk[k].Tags[1]),
					[]byte(mk[k].Tags[2]),
				}
				joinedVal := bytes.Join(bt, []byte("_"))
				err = bucket.Put([]byte(k), joinedVal)

				if err != nil {
					return fmt.Errorf("Failed to insert '%s': '%v'", k, err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("bbolt db.Update in DBInit failed '%v'", err)
	}
	return nil
}

func DBInsert(db *bolt.DB, k string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("TestWords"))
		if err != nil {
			return fmt.Errorf("Failed to create bucket: %v", err)
		}
		if bucket.Get([]byte(k)) == nil { // don't override key/val if exists
			err = bucket.Put([]byte(k), []byte(""))
			if err != nil {
				return fmt.Errorf("Failed to insert '%s': '%v'", k, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("bbolt db.Update in DBInsert failed '%v'", err)
	}
	return nil
}

func DBView(db *bolt.DB, find string) ([]byte, error) {
	var result []byte
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("TestWords"))
		if bucket == nil {
			return fmt.Errorf("Failed to find bucket")
		}
		result = bucket.Get([]byte(find))
		if result == nil {
			println(bucket.Get([]byte(find)))
		}
		return nil
	})
	if err != nil {
		return result, fmt.Errorf("bbolt db.View in DBView failed '%v'", err)
	}
	return result, nil
}
