// Adapted from https://github.com/grahamrhay/go-ping/blob/master/datastore.go (MIT Licensed) - (c) Graham Hay
package main

import (
	"github.com/steveyen/gkvlite"
	"log"
	"os"
)

type Store struct {
	f *os.File
	s *gkvlite.Store
}

func openStore() (*Store, error) {
	f, err := os.OpenFile("./db", os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return nil, err
	}
	s, err := gkvlite.NewStore(f)
	if err != nil {
		return nil, err
	}
	return &Store{f: f, s: s}, nil
}

func deferCloseStore(store *Store) {
	//defer store.s.Close()
	//defer store.f.Close()
}

func writeToStore(store *Store, coll string, key string, item string) error {
	log.Printf("Writing item to store. Coll: %v, Key: %v, Item: %v\n", coll, key, item)
	c := store.s.GetCollection(coll)
	if c == nil {
		log.Println("Collection doesn't exist, creating it")
		c = store.s.SetCollection(coll, nil)
	}
	err := c.Set([]byte(key), []byte(item))
	if err != nil {
		return err
	}
	err1 := store.s.Flush()
	if err1 != nil {
		return err1
	}
	return store.f.Sync()
}

func getFromStore(store *Store, coll string, key string) (string, error) {
	log.Printf("Retrieving item from store. Coll: %v, Key: %v\n", coll, key)
	c := store.s.GetCollection(coll)
	if c == nil {
		log.Println("Collection doesn't exist")
		return "", nil
	}
	itemBytes, err := c.Get([]byte(key))
	if err != nil {
		return "", err
	}
	if itemBytes == nil {
		return "", nil
	}
	item := string(itemBytes[:])
	return item, nil
}
