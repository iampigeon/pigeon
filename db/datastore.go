package db

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/iampigeon/pigeon"

	"github.com/boltdb/bolt"
)

// Datastore sote data in db using bolt as a db backend
type Datastore struct {
	DB *bolt.DB
}

var (
	msgBucket      = []byte("messages")
	subjectsBucket = []byte("subjects")
)

// NewDatastore returns a new datastore instance or an error if
// a datasore cannot be returned
func NewDatastore(path string) (*Datastore, error) {
	db, err := bolt.Open(path, os.ModePerm, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, berr := tx.CreateBucketIfNotExists(msgBucket)
		return berr
	})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, berr := tx.CreateBucketIfNotExists(subjectsBucket)
		return berr
	})
	if err != nil {
		return nil, err
	}

	return &Datastore{DB: db}, nil
}

func getMock() (*pigeon.Mock, error) {
	data, err := ioutil.ReadFile("./data.json")
	if err != nil {
		return nil, err
	}

	mock := new(pigeon.Mock)
	err = json.Unmarshal(data, &mock)
	if err != nil {
		return nil, err
	}

	return mock, nil
}
