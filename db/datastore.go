package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/iampigeon/pigeon"

	arango "github.com/arangodb/go-driver"
	"github.com/boltdb/bolt"
)

const (
	pigeonDB = "pigeon_db"
)

// Datastore sote data in db using bolt as a db backend
type Datastore struct {
	DB           *bolt.DB
	ArangoClient arango.Client
	PigeonDB     arango.Database
	Context      *context.Context
}

var (
	msgBucket      = []byte("messages")
	subjectsBucket = []byte("subjects")
)

// NewDatastore returns a new datastore instance or an error if
// a datasore cannot be returned
func NewDatastore(path string, conn arango.Connection) (*Datastore, error) {
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

	//ArangoClient ...
	cl, err := arango.NewClient(arango.ClientConfig{
		Connection: conn,
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("connection arango")
	fmt.Println(cl)

	ctx := context.Background()
	found, err := cl.DatabaseExists(ctx, pigeonDB)
	if err != nil {
		fmt.Println("88888888")
		fmt.Println(err)
		return nil, err
	}

	var pdb arango.Database
	if !found {
		opt := new(arango.CreateDatabaseOptions)
		pdb, err = cl.CreateDatabase(ctx, pigeonDB, opt)
		if err != nil {
			fmt.Println("333333")
			return nil, err
		}
		return &Datastore{
			DB:           db,
			ArangoClient: cl,
			PigeonDB:     pdb,
			Context:      &ctx,
		}, nil
	}
	pdb, err = cl.Database(ctx, pigeonDB)
	if err != nil {
		fmt.Println("444444444")
		return nil, err
	}

	fmt.Println("5555555555")
	return &Datastore{
		DB:           db,
		ArangoClient: cl,
		PigeonDB:     pdb,
		Context:      &ctx,
	}, nil
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
