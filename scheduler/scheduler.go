package scheduler

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/iampigeon/pigeon"
	"github.com/iampigeon/pigeon/db"
	pb "github.com/iampigeon/pigeon/proto"
	"github.com/oklog/ulid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// TODO(ja): remove this struct.

// StorageConfig is a struct that will be deleted.
type StorageConfig struct {
	// BoltDatabase     string        // File to use as bolt database.
	RedisURL         string        // URL of the redis server
	RedisLog         bool          // log database commands
	RedisMaxIdle     int           // maximum number of idle connections in the pool
	RedisDatabase    int           // redis database to use
	RedisIdleTimeout time.Duration // timeout for idle connections

	MessageStore *db.MessageStore
}

// New builds a new pigeon.Store backed by bolt DB.
//
// In case of any error it panics.
func New(config StorageConfig) pigeon.SchedulerService {
	s := &service{
		pq:  newPriorityQueue(config),
		idc: make(chan ulid.ULID),

		ms: config.MessageStore,
	}

	go s.run()

	return s
}

var msgBucket = []byte("messages")

type service struct {
	// db *bolt.DB
	pq *priorityQueue

	idc chan ulid.ULID

	ms *db.MessageStore
}

func (s *service) Put(id ulid.ULID, content []byte, endpoint pigeon.NetAddr, status pigeon.MessageStatus, subjectID string) error {
	// TODO(ja): use secure connections

	host, port, err := net.SplitHostPort(string(endpoint))
	if err != nil {
		return err
	}

	endpoint = pigeon.NetAddr(net.JoinHostPort(host, port))
	log.Println(endpoint)

	conn, err := grpc.Dial(string(endpoint), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewBackendServiceClient(conn)
	resp, err := client.Approve(context.Background(), &pb.ApproveRequest{Content: content})
	if err != nil {
		// update status to crashed-approve
		e := s.ms.UpdateStatus(id, pigeon.StatusCrashedApprove)
		if e != nil {
			return e
		}

		// send http error via pigeon-htpp
		err := s.sendCallbackHTTPMessage(subjectID, "could not deliver message")
		if err != nil {
			// TODO(ca): check this error
			log.Printf("Error: could not send callback http message %v", err)
			return err
		}

		return err
	}
	if !resp.Valid {
		// update status to failed-approve
		e := s.ms.UpdateStatus(id, pigeon.StatusFailedApprove)
		if e != nil {
			return e
		}

		// send http error via pigeon-htpp
		err := s.sendCallbackHTTPMessage(subjectID, "could not deliver message")
		if err != nil {
			return err
		}

		if resp.Error != nil {
			return errors.Errorf("invalid message, %s", resp.Error.Message)
		}
		return errors.New("invalid message")
	}

	m := pigeon.Message{
		ID:        id,
		Content:   content,
		Endpoint:  endpoint,
		Status:    status,
		SubjectID: subjectID,
	}

	err = s.ms.AddMessage(m)
	if err != nil {
		return err
	}

	s.idc <- id

	return nil
}

func (s *service) Get(id ulid.ULID) (*pigeon.Message, error) {
	msg, err := s.ms.GetMessage(id)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *service) Update(id ulid.ULID, content []byte) error {
	err := s.ms.UpdateContent(id, content)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) Cancel(id ulid.ULID) error {
	ok, err := s.pq.DeleteByID(id)
	if err != nil {
		return err
	}

	if !ok {
		log.Printf("%s not found in priority queue", id)
		return nil
	}

	err = s.ms.UpdateStatus(id, pigeon.StatusCancelled)
	if err != nil {
		return err
	}

	return nil
}

// Run in its goroutine
func (s *service) run() {
	var next uint64
	var timer *time.Timer

	pq := s.pq
	for {
		var tick <-chan time.Time

		top := pq.Peek()
		if top != nil {
			if t := top.Time(); t < next || next == 0 {
				var delay int64
				now := ulid.Timestamp(time.Now())
				if t >= now {
					delay = int64(t - now)
				}

				if timer == nil {
					timer = time.NewTimer(time.Duration(delay) * time.Millisecond)
				} else {
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer = time.NewTimer(time.Duration(delay) * time.Millisecond)
				}
			}
		}

		if timer != nil && top != nil {
			tick = timer.C
		}

		select {
		case <-tick:
			id, err := pq.Pop()
			if err != nil {
				log.Printf(err.Error())
			}

			if id != nil {
				go s.send(*id)
			}
			next = 0
		case id := <-s.idc:
			pq.Push(id)
		}
	}
}

func (s *service) send(id ulid.ULID) {
	msg, err := s.Get(id)
	if err != nil {
		log.Printf("Error: could not get message %s, %v", id, err)
		return
	}

	// TODO(ja): use secure connections
	conn, err := grpc.Dial(string(msg.Endpoint), grpc.WithInsecure())
	if err != nil {
		log.Printf("Error: could not connect to backend at %s, %v", msg.Endpoint, err)
		return
	}
	defer conn.Close()

	client := pb.NewBackendServiceClient(conn)
	resp, err := client.Deliver(context.Background(), &pb.DeliverRequest{Content: msg.Content})
	if err != nil {
		log.Printf("Error: could not deliver message %s, %v", msg.ID, err)

		// update status to crashed-deliver
		e := s.ms.UpdateStatus(id, pigeon.StatusCrashedDeliver)
		if e != nil {
			log.Printf("Error: could not update message status %s, %v", msg.ID, err)
			return
		}

		// send http error via pigeon-htpp
		err := s.sendCallbackHTTPMessage(msg.SubjectID, "could not deliver message")
		if err != nil {
			// TODO(ca): check this error
			log.Printf("Error: could not send callback http message %v", err)
			return
		}

		return
	}
	if resp.Error != nil {
		log.Printf("Error: failed to deliver message %s, %v", msg.ID, resp.Error.Message)

		// update status to failed-deliver
		e := s.ms.UpdateStatus(id, pigeon.StatusFailedDeliver)
		if e != nil {
			// TODO(ca): check this error
			log.Printf("Error: could not update message status %s, %v", msg.ID, err)
			return
		}

		// send http error via pigeon-htpp
		err := s.sendCallbackHTTPMessage(msg.SubjectID, "failed to deliver message")
		if err != nil {
			// TODO(ca): check this error
			log.Printf("Error: could not send callback http message %v", err)
			return
		}

		return
	}

	e := s.ms.UpdateStatus(id, pigeon.StatusSent)
	if e != nil {
		log.Printf("Error: could not update message status %s, %v", msg.ID, err)
		return
	}
}

func (s *service) sendCallbackHTTPMessage(subjectID string, messageError string) error {
	id, err := generateID(0)
	if err != nil {
		return err
	}

	body := pigeon.Response{
		Error: messageError,
	}

	c, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// TODO(ca): use callback_post_url as a new HTTP message
	err = s.Put(*id, c, pigeon.ServicePigeonHTTP, pigeon.StatusPending, subjectID)
	if err != nil {
		return err
	}

	return nil
}

// TODO(ca): move this to other site.
func generateID(criteriaDelay time.Duration) (*ulid.ULID, error) {
	delay := criteriaDelay

	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	id, err := ulid.New(
		ulid.Timestamp(time.Now().Add(delay)),
		entropy,
	)
	if err != nil {
		//TODO: move this message
		log.Println("Failed to create message id, %v", err)
		return nil, err
	}

	return &id, nil
}
