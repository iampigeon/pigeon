package pigeon

import (
	"net/url"

	"github.com/oklog/ulid"
)

// TODO: document this
const (
	// StatusPending ...
	StatusPending = "pending"
	// StatusSent ...
	StatusSent = "sent"
	// StatusFailedApprove ...
	StatusFailedApprove = "failed-approve"
	// StatusCrashedApprove ...
	StatusCrashedApprove = "crashed-approve"
	// StatusFailedDeliver ...
	StatusFailedDeliver = "failed-deliver"
	// StatusCrashedDeliver ...
	StatusCrashedDeliver = "crashed-deliver"
	// StatusCancelled ...
	StatusCancelled = "cancelled"

	// EndpointMQTT ...
	EndpointMQTT = "pigeon-mqtt:9010"
	// EndpointHTTP ...
	EndpointHTTP = "pigeon-http:9020"

	// ServicePigeonMQTT ...
	ServicePigeonMQTT = "mqtt"
	// ServicePigeonSMS ...
	ServicePigeonSMS = "sms"
	// ServicePigeonHTTP ...
	ServicePigeonHTTP = "http"
)

// NetAddr is the network address of the Backend service where to validate and
// send messages.
type NetAddr string

// MessageStatus ...
type MessageStatus string

// Message describes a message that needs to be delivered by the system.
type Message struct {
	// ID is an ULID that uniquely identifies (https://github.com/alizain/ulid)
	// a message and encodes the time when the message needs to be sent.
	ID ulid.ULID `json:"id", arango:"id"`

	// Content is an arbitrary byte slice that describes the message to
	// be sent.
	//
	// The format of the content varies by the Backend used, and to avoid
	// latter failures the Backend must validate the content before the
	// approval of the message.
	Content []byte `json:"content,string", arango:"content"`

	// Endpoint identifies the Backend service used to send the message.
	Endpoint NetAddr `json:"-", arango:"endpoint"`

	// Status ...
	Status MessageStatus `json:"status", arango:"status"`

	SubjectID string `json:"subject_id", arango:"subject_id"`
	UserID    string `json:"-", arango:"user_id"`

	// Subject virtual reference to subject
	Subject *Subject `json:"-"`
}

// SchedulerService stores and keep track of the statuses of messages.
type SchedulerService interface {
	// Put stores a message content and schedule the delivery on t time.
	// TODO(ca): change subjectID params to ulid.ULID type
	Put(id ulid.ULID, content []byte, endpoint NetAddr, status MessageStatus, subjectID, userID string) error

	// Get retrieves the message with the given id.
	//
	// In case of any error the Message will be nil.
	Get(id ulid.ULID, u *User) (*Message, error)
	GetMessageByID(id ulid.ULID) (*Message, error)

	// Update updates the content of the message with the given id.
	Update(id ulid.ULID, content []byte) error

	// Cancel cancel the message with the given id.
	Cancel(id ulid.ULID) error
}

// Backend manages the approval and delivery of messages.
type Backend interface {
	// Aprove validates the content of a message.
	//
	// If the message is valid the error will be nil, otherwise the error
	// must be non-nil and describe why the message is invalid.
	Approve(content []byte) (ok bool, err error)

	// Deliver delivers the message encoded in content.
	Deliver(content []byte) error
}

// Subject ...
type Subject struct {
	ID       string            `json:"id"`
	UserID   string            `json:"user_id"`
	Name     string            `json:"name"`
	Channels []*SubjectChannel `json:"channels"`

	User *User `json:"-"`
}

// SubjectChannel ...
type SubjectChannel struct {
	ID             string                 `json:"id"`
	ChannelID      string                 `json:"channel_id"`
	Options        map[string]interface{} `json:"options"`
	CriteriaID     string                 `json:"criteria_id"`
	CriteriaCustom int64                  `json:"criteria_custom"`

	Channel  *Channel  `json:"-"`
	Criteria *Criteria `json:"-"`
}

// TODO: move to respective pigeon repository and get from package
// MQTTContent ...
type MQTTContent struct {
	Payload map[string]interface{} `json:"mqtt_payload"`
}

// MQTTOptions ...
type MQTTOptions struct {
	Topic string `json:"mqtt_topic"`
}

// MQTT ...
type MQTT struct {
	Topic   string                 `json:"mqtt_topic"`
	Payload map[string]interface{} `json:"mqtt_payload"`
}

// Message message struct
// HTTP ...
type HTTP struct {
	URL     *url.URL               `json:"url"`
	Body    string                 `json:"body,omitempty"`
	Headers map[string]interface{} `json:"headers,omitempty"`
}

// HTTPContent ...
type HTTPContent struct {
	Body string `json:"body,omitempty"`
}

// HTTPOptions ...
type HTTPOptions struct {
	URL     *url.URL               `json:"url"`
	Headers map[string]interface{} `json:"headers,omitempty"`
}

// TODO: move to respective pigeon repository and get from package
// SMS ...
type SMS struct {
	Phone string `json:"phone"`
	Text  string `json:"text"`
}

// Push ...
type Push struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Token string `json:"token"`
}

// PushContent ...
type PushContent struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Token string `json:"token"`
}

// PushOptions ...
type PushOptions struct{}

// Channel ...
type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Host string `json:"host"`
}

// User ...
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	APIKey   string `json:"api_key"`
}

// Criteria ...
type Criteria struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

// Mock ...
type Mock struct {
	Users     []*User     `json:"users"`
	Channels  []*Channel  `json:"channels"`
	Subjects  []*Subject  `json:"subjects"`
	Criterias []*Criteria `json:"criterias"`
	// SubjectsChannels []*SubjectsChannels `json:"subjects_channels"`
	//Messages         []*Message          `json:"messages"`
}

// Response ...
type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Meta  interface{} `json:"meta,omitempty"`
	Error interface{} `json:"error,omitempty"`
}
