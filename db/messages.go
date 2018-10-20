package db

import (
	"context"
	"fmt"
	"log"

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
		"user_id":    m.UserID,
	}

	_, err := ss.Collection.CreateDocument(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}

// GetMessage ...
func (ss *MessageStore) GetMessage(id ulid.ULID, u *pigeon.User) (*pigeon.Message, error) {
	var msg pb.Message

	query := fmt.Sprintf(`
	FOR m IN message_collection
	FILTER m.id == '%s'
	FILTER m.user_id == '%s'
	RETURN m
	`, id.String(), u.ID)
	fmt.Println(query)

	cursor, err := ss.Collection.Database().Query(*ss.Dst.Context, query, nil)
	if err != nil {
		return nil, err
	}

	meta, err := cursor.ReadDocument(*ss.Dst.Context, &msg)
	if err != nil {
		return nil, err
	}
	log.Println("META DATA \n", meta)

	return &pigeon.Message{
		ID:        id,
		Content:   msg.Content,
		Endpoint:  pigeon.NetAddr(msg.Endpoint),
		Status:    pigeon.MessageStatus(msg.Status),
		SubjectID: msg.SubjectId,
	}, nil
}

// GetMessageByID ...
func (ss *MessageStore) GetMessageByID(id ulid.ULID) (*pigeon.Message, error) {
	var msg pb.Message

	query := fmt.Sprintf(`
	FOR m IN message_collection
	FILTER m.id == '%s'
	RETURN m
	`, id.String())
	fmt.Println(query)

	cursor, err := ss.Collection.Database().Query(*ss.Dst.Context, query, nil)
	if err != nil {
		return nil, err
	}

	meta, err := cursor.ReadDocument(*ss.Dst.Context, &msg)
	if err != nil {
		return nil, err
	}
	log.Println("META DATA \n", meta)

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
	query := fmt.Sprintf(`
	FOR msg IN message_collection
	FILTER msg.id == '%s'
	UPDATE msg WITH { _key: msg._key,  content: '%s' }
	IN message_collection
	`, id.String(), content)
	fmt.Println(query)

	_, err := ss.Collection.Database().Query(*ss.Dst.Context, query, nil)
	if err != nil {
		return err
	}

	return nil
}

// UpdateStatus ...
func (ss *MessageStore) UpdateStatus(id ulid.ULID, status pigeon.MessageStatus) error {
	query := fmt.Sprintf(`
	FOR msg IN message_collection
	FILTER msg.id == '%s'
	UPDATE msg WITH { _key: msg._key,  status: '%s' }
	IN message_collection
	`, id.String(), string(status))

	_, err := ss.Collection.Database().Query(*ss.Dst.Context, query, nil)
	if err != nil {
		return err
	}

	return nil
}
