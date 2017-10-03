package pigeon

import (
	"github.com/oklog/ulid"
)

// NetAddr is the network address of the Backend service where to validate and
// send messages.
type NetAddr string

// Message describes a message that needs to be delivered by the system.
type Message struct {
	// ID is an ULID that uniquely identifies (https://github.com/alizain/ulid)
	// a message and encodes the time when the message needs to be sent.
	ID ulid.ULID

	// Content is an arbitrary byte slice that describes the message to
	// be sent.
	//
	// The format of the content varies by the Backend used, and to avoid
	// latter failures the Backend must validate the content before the
	// approval of the message.
	Content []byte

	// Endpoint identifies the Backend service used to send the message.
	Endpoint NetAddr
}

// SchedulerService stores and keep track of the statuses of messages.
type SchedulerService interface {
	// Put stores a message content and schedule the delivery on t time.
	Put(id ulid.ULID, content []byte, endpoint NetAddr) error

	// Get retrieves the message with the given id.
	//
	// In case of any error the Message will be nil.
	Get(id ulid.ULID) (*Message, error)

	// Update updates the content of the message with the given id.
	Update(id ulid.ULID, content []byte) error
}

// Backend manages the approval and delivery of messages.
type Backend interface {
	// Aprove validates the content of a message.
	//
	// If the message is valid the error will be nil, otherwise the error
	// must be non-nil and describe why the message is invalid.
	Aprove(content []byte) (ok bool, err error)

	// Deliver delivers the message encoded in content.
	Deliver(content []byte) error
}
