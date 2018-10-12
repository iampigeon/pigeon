package db

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	"github.com/iampigeon/pigeon"
	"github.com/oklog/ulid"

	pb "github.com/iampigeon/pigeon/proto"
)

// MessageStore ...
type MessageStore struct {
	*Datastore
}

// GetMessages ...
// func (ss *MessageStore) GetMessages() ([]*pigeon.Message, error) {
// 	data, err := ioutil.ReadFile("../data.json")
// 	if err != nil {
// 		return nil, err
// 	}

// 	var mock pigeon.Mock
// 	err = json.Unmarshal(data, &mock)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return mock.Messages, nil
// }

// AddMessage ...
func (ss *MessageStore) AddMessage(m pigeon.Message) error {
	err := ss.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(msgBucket)

		k, merr := m.ID.MarshalBinary()
		if merr != nil {
			return merr
		}
		v, jerr := proto.Marshal(&pb.Message{
			Id:       m.ID.String(),
			Content:  m.Content,
			Endpoint: string(m.Endpoint),
			Status:   string(m.Status),
		})
		if jerr != nil {
			return jerr
		}
		return b.Put(k, v)
	})
	if err != nil {
		return err
	}

	fmt.Println("se guardo")

	return nil
}

// GetMessage ...
func (ss *MessageStore) GetMessage(id ulid.ULID) (*pigeon.Message, error) {
	var msg pb.Message

	err := ss.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(msgBucket)
		k, err := id.MarshalBinary()
		if err != nil {
			return err
		}
		v := b.Get(k)
		if err := proto.Unmarshal(v, &msg); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pigeon.Message{
		ID:       id,
		Content:  msg.Content,
		Endpoint: pigeon.NetAddr(msg.Endpoint),
		Status:   pigeon.MessageStatus(msg.Status),
	}, nil
}

// UpdateContent ...
func (ss *MessageStore) UpdateContent(id ulid.ULID, content []byte) error {
	var msg pb.Message

	err := ss.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(msgBucket)
		k, err := id.MarshalBinary()
		if err != nil {
			return err
		}
		v := b.Get(k)
		if err = proto.Unmarshal(v, &msg); err != nil {
			return err
		}
		msg.Content = content
		v, err = proto.Marshal(&msg)
		if err != nil {
			return err
		}
		return b.Put(k, v)
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateStatus ...
func (ss *MessageStore) UpdateStatus(id ulid.ULID, status pigeon.MessageStatus) error {
	var msg pb.Message

	err := ss.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(msgBucket)
		k, err := id.MarshalBinary()
		if err != nil {
			return err
		}
		v := b.Get(k)
		if err = proto.Unmarshal(v, &msg); err != nil {
			return err
		}
		msg.Status = string(status)
		v, err = proto.Marshal(&msg)
		if err != nil {
			return err
		}
		return b.Put(k, v)
	})
	if err != nil {
		return err
	}

	return nil
}
