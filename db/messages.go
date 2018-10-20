package db

import (
	"context"

	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	"github.com/iampigeon/pigeon"
	"github.com/oklog/ulid"

	arango "github.com/arangodb/go-driver"
	pb "github.com/iampigeon/pigeon/proto"
)

const (
	msgCollection = "message_collection"
)

// MessageStore ...
type MessageStore struct {
	Dst        *Datastore
	Collection arango.Collection
}

// NewMessageStore ...
func NewMessageStore(dst *Datastore) (*MessageStore, error) {
	found, err := dst.PigeonDB.CollectionExists(*dst.Context, msgCollection)
	if err != nil {
		return nil, err
	}

	var col arango.Collection

	if !found {
		opt := new(arango.CreateCollectionOptions)
		col, err = dst.PigeonDB.CreateCollection(*dst.Context, msgCollection, opt)
		if err != nil {
			return nil, err
		}

		return &MessageStore{
			Dst:        dst,
			Collection: col,
		}, nil
	}

	col, err = dst.PigeonDB.Collection(*dst.Context, msgCollection)
	if err != nil {
		return nil, err
	}

	return &MessageStore{
		Dst:        dst,
		Collection: col,
	}, nil
}

// AddMessage ...
func (ss *MessageStore) AddMessage(m pigeon.Message) error {
	ctx := context.Background()
	msg := map[string]interface{}{
		"id":         m.ID.String(),
		"content":    m.Content,
		"endpoint":   string(m.Endpoint),
		"status":     string(m.Status),
		"subject_id": string(m.SubjectID),
	}

	_, err := ss.Collection.CreateDocument(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}

// GetMessage ...
func (ss *MessageStore) GetMessage(id ulid.ULID) (*pigeon.Message, error) {
	var msg pb.Message

	err := ss.Dst.DB.View(func(tx *bolt.Tx) error {
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
		ID:        id,
		Content:   msg.Content,
		Endpoint:  pigeon.NetAddr(msg.Endpoint),
		Status:    pigeon.MessageStatus(msg.Status),
		SubjectID: msg.SubjectId,
	}, nil
}

// UpdateContent ...
func (ss *MessageStore) UpdateContent(id ulid.ULID, content []byte) error {
	var msg pb.Message

	err := ss.Dst.DB.Update(func(tx *bolt.Tx) error {
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

	err := ss.Dst.DB.Update(func(tx *bolt.Tx) error {
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
