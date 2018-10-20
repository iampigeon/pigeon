package httpsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/WiseGrowth/go-wisebot/logger"
	"github.com/iampigeon/pigeon"
	"github.com/iampigeon/pigeon/db"
	"github.com/iampigeon/pigeon/proto"
	"github.com/julienschmidt/httprouter"
	"github.com/oklog/ulid"
	"github.com/urfave/negroni"
	"google.golang.org/grpc"
)

type key int

const (
	// TODO: get port from other side
	httpPort = 9000
	grpcPort = 9001

	loggerKey key = iota

	// TODO: check this
	APIKey = "12345"
)

// Subject ...
type Subject struct {
	Name     string   `json:"name"`
	Channels []string `json:"channels"`
}

// SubjectsResponse ...
type SubjectsResponse struct {
	Subjects []*Subject `json:"subjects"`
}

// Response ...
type Response struct {
	Data interface{} `json:"data,omitempty"`
	Meta interface{} `json:"meta,omitempty"`
}

// MessageRequestChannels ...
//type MessageRequestChannels struct {
//	MQTT *pigeon.MQTT `json:"mqtt,omitempty"`
//	SMS  *pigeon.SMS  `json:"sms,omitempty"`
//	HTTP *pigeon.HTTP `json:"http,omitempty"`
//	Push *pigeon.Push `json:"push,omitempty"`
//}

// MessageRequest ...
type MessageRequest struct {
	Message *struct {
		SubjectName string                 `json:"subject_name"`
		Channels    map[string]interface{} `json:"channels"`
	} `json:"message"`
}

// MessagesResponse ...
type MessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
}

// MessageStatusResponse ...
type MessageStatusResponse struct {
	Status string `json:"status"`
}

// MessageByIDResponse ...
type MessageByIDResponse struct {
	Message *pigeon.Message `json:"message"`
}

// MessageCancelResponse ...
type MessageCancelResponse struct {
	Status string `json:"status"`
}

// MessageResponse ...
type MessageResponse struct {
	ID      string `json:"id,omitempty"`
	Channel string `json:"channel,omitempty"`
	Error   string `json:"error,omitempty"`
}

type getSubjectsContext struct {
	SubjectStore *db.SubjectStore
	UserStore    *db.UserStore
	ChannelStore *db.ChannelStore
}

type getMessageStatusContext struct {
	MessageStore *db.MessageStore
	UserStore    *db.UserStore
	SubjectStore *db.SubjectStore
}

type postMessageContext struct {
	SubjectStore  *db.SubjectStore
	UserStore     *db.UserStore
	ChannelStore  *db.ChannelStore
	CriteriaStore *db.CriteriaStore
}

type postCancelMessageContext struct {
	MessageStore *db.MessageStore
	SubjectStore *db.SubjectStore
	UserStore    *db.UserStore
}

type getMessageByIDContext struct {
	MessageStore *db.MessageStore
	SubjectStore *db.SubjectStore
	UserStore    *db.UserStore
}

// NewHTTPServer returns an initialized server
//
// GET /api/v1/subjects
// POST /api/v1/messages
// GET /api/v1/messages/:id
// GET /api/v1/messages/:id/status
// POST /api/v1/messages/:id/cancel
//
func NewHTTPServer(database *db.Datastore) *http.Server {
	router := httprouter.New()

	// stores
	ss := &db.SubjectStore{database}
	us := &db.UserStore{database}
	cs := &db.ChannelStore{database}
	ms := &db.MessageStore{database}
	ts := &db.CriteriaStore{database}

	router.GET("/api/v1/subjects", getSubjectsHTTPHandler(getSubjectsContext{SubjectStore: ss, UserStore: us, ChannelStore: cs}))
	router.GET("/api/v1/messages/:id", getMessageByIDHTTPHandler(getMessageByIDContext{UserStore: us, SubjectStore: ss, MessageStore: ms}))
	router.POST("/api/v1/messages", postMessageHTTPHandler(postMessageContext{UserStore: us, SubjectStore: ss, ChannelStore: cs, CriteriaStore: ts}))
	router.GET("/api/v1/messages/:id/status", getStatusMessageHTTPHandler(getMessageStatusContext{MessageStore: ms, UserStore: us}))
	router.POST("/api/v1/messages/:id/cancel", postCancelMessageHTTPHandler(postCancelMessageContext{MessageStore: ms, UserStore: us, SubjectStore: ss}))

	addr := fmt.Sprintf(":%d", httpPort)
	routes := negroni.Wrap(router)
	n := negroni.New(negroni.HandlerFunc(httpLogginMiddleware), routes)
	server := &http.Server{Addr: addr, Handler: n}

	return server
}

func getSubjectsHTTPHandler(ctx getSubjectsContext) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")

		// get api key from header
		apiKey := r.Header.Get("X-Api-Key")

		// get user by api key
		user, err := ctx.UserStore.GetUserByAPIKey(apiKey)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// define http response
		subjectsResponse := new(SubjectsResponse)
		subjectsResponse.Subjects = make([]*Subject, 0)

		// get user subjects
		subjects, err := ctx.SubjectStore.GetSubjectsByUserID(user.ID)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		for _, subject := range subjects {
			s := new(Subject)
			s.Name = subject.Name
			s.Channels = make([]string, 0)

			// get channels
			for _, channel := range subject.Channels {
				c, e := ctx.ChannelStore.GetChannelById(channel.ChannelID)
				if e != nil {
					getLogger(r).Error(e)
				} else {
					s.Channels = append(s.Channels, c.Name)
				}
			}

			subjectsResponse.Subjects = append(subjectsResponse.Subjects, s)
		}

		response := new(Response)
		response.Data = subjectsResponse

		if err := json.NewEncoder(w).Encode(response); err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func getMessageByIDHTTPHandler(ctx getMessageByIDContext) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")

		// get api key from header
		apiKey := r.Header.Get("X-Api-Key")

		// get user by api key
		_, err := ctx.UserStore.GetUserByAPIKey(apiKey)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// define id
		id, err := ulid.Parse(ps.ByName("id"))
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// get message by id
		msg, err := ctx.MessageStore.GetMessage(id)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// // Get subject
		// subject, err := ctx.SubjectStore.GeSubjectByID(msg.SubjectID)
		// if err != nil {
		// 	getLogger(r).Error(err)
		// 	http.Error(w, err.Error(), http.StatusBadRequest)
		// 	return
		// }

		// // Validate if message belongs to user
		// if subject.UserID != user.ID {
		// 	getLogger(r).Error(err)
		// 	http.Error(w, err.Error(), http.StatusForbidden)
		// 	return
		// }

		// prepare response
		messageByIDResponse := new(MessageByIDResponse)
		messageByIDResponse.Message = msg

		response := new(Response)
		response.Data = messageByIDResponse

		if err := json.NewEncoder(w).Encode(response); err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func postMessageHTTPHandler(ctx postMessageContext) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

		w.Header().Set("Content-Type", "application/json")

		// get user by api key
		apiKey := r.Header.Get("X-Api-Key")
		user, err := ctx.UserStore.GetUserByAPIKey(apiKey)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Bind and parse request/json
		payload := new(MessageRequest)

		// Read body and define string
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// Decode body string to payload
		err = json.Unmarshal(body, payload)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check and get subject id from mqtt payload key
		subject, err := ctx.SubjectStore.GetUserSubjectByName(user.ID, payload.Message.SubjectName)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO(ca): check this
		addr := fmt.Sprintf("localhost:%d", grpcPort)

		//grpc connection
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		// Define scheduler proto client
		client := proto.NewSchedulerServiceClient(conn)

		// Prepare message array response
		messagesResponses := new(MessagesResponse)

		for channelName, channelValue := range payload.Message.Channels {
			response := new(MessageResponse)
			response.Channel = channelName

			// get subjectChannel according current channel name
			subjectChannel, err := getSubjectChannelByName(channelName, subject, ctx.ChannelStore)
			if err != nil {
				response.Error = err.Error()
				messagesResponses.Messages = append(messagesResponses.Messages, *response)
				continue
			}

			criteriaDelay, err := ctx.CriteriaStore.GetCriteriaDelay(subjectChannel.CriteriaID, subjectChannel.CriteriaCustom)
			if err != nil {
				response.Error = err.Error()
				messagesResponses.Messages = append(messagesResponses.Messages, *response)
				continue
			}

			// prepare and send pigeon-mqtt message
			if channelName == pigeon.ServicePigeonMQTT {
				// define mqtt content
				var mqttContent *pigeon.MQTTContent

				c, err := json.Marshal(channelValue)
				if err != nil {
					response.Error = fmt.Sprintf("invalid json mashall for %s channel", channelName)
					messagesResponses.Messages = append(messagesResponses.Messages, *response)
					continue
				}
				err = json.Unmarshal(c, &mqttContent)
				if err != nil {
					response.Error = fmt.Sprintf("invalid cast parse for %s channel", channelName)
					messagesResponses.Messages = append(messagesResponses.Messages, *response)
					continue
				}

				//define mqtt options
				var mqttOptions *pigeon.MQTTOptions

				c, err = json.Marshal(subjectChannel.Options)
				if err != nil {
					response.Error = fmt.Sprintf("invalid json mashall for %s channel", channelName)
					messagesResponses.Messages = append(messagesResponses.Messages, *response)
					continue
				}
				err = json.Unmarshal(c, &mqttOptions)
				if err != nil {
					response.Error = fmt.Sprintf("invalid cast parse for %s channel", channelName)
					messagesResponses.Messages = append(messagesResponses.Messages, *response)
					continue
				}

				//define mqtt
				mqtt := pigeon.MQTT{
					Topic:   mqttOptions.Topic,
					Payload: mqttContent.Payload,
				}

				content, err := json.Marshal(mqtt)
				if err != nil {
					response.Error = fmt.Sprintf("invalid json mashall for %s channel", channelName)
					messagesResponses.Messages = append(messagesResponses.Messages, *response)
					continue
				}

				// 	Send MQTT message
				id, err := sendMessage(client, string(content), pigeon.EndpointMQTT, subject.ID, criteriaDelay)
				if err != nil {
					response.Error = err.Error()
				} else {
					// Save id inside message response
					response.ID = id
				}

				messagesResponses.Messages = append(messagesResponses.Messages, *response)

				// prepare and send pigeon-http message
			} else if channelName == pigeon.ServicePigeonHTTP {
				// define http content
				var httpContent *pigeon.HTTPContent

				c, err := json.Marshal(channelValue)
				if err != nil {
					response.Error = fmt.Sprintf("invalid json mashall for %s channel", channelName)
					messagesResponses.Messages = append(messagesResponses.Messages, *response)
					continue
				}
				err = json.Unmarshal(c, &httpContent)
				if err != nil {
					response.Error = fmt.Sprintf("invalid cast parse for %s channel", channelName)
					messagesResponses.Messages = append(messagesResponses.Messages, *response)
					continue
				}

				//define http options
				var httpOptions *pigeon.HTTPOptions

				c, err = json.Marshal(subjectChannel.Options)
				if err != nil {
					response.Error = fmt.Sprintf("invalid json mashall for %s channel", channelName)
					messagesResponses.Messages = append(messagesResponses.Messages, *response)
					continue
				}
				err = json.Unmarshal(c, &httpOptions)
				if err != nil {
					response.Error = fmt.Sprintf("invalid cast parse for %s channel", channelName)
					messagesResponses.Messages = append(messagesResponses.Messages, *response)
					continue
				}

				//define http
				http := pigeon.HTTP{
					Headers: httpOptions.Headers,
					URL:     httpOptions.URL,
					Body:    httpContent.Body,
				}

				content, err := json.Marshal(http)
				if err != nil {
					response.Error = fmt.Sprintf("invalid json mashall for %s channel", channelName)
					messagesResponses.Messages = append(messagesResponses.Messages, *response)
					continue
				}

				// 	Send HTTP message
				id, err := sendMessage(client, string(content), pigeon.EndpointHTTP, subject.ID, criteriaDelay)
				if err != nil {
					response.Error = err.Error()
				} else {
					// Save id inside message response
					response.ID = id
				}

				messagesResponses.Messages = append(messagesResponses.Messages, *response)
			} else {
				response.Error = fmt.Sprintf("invalid channel name %s", channelName)
				messagesResponses.Messages = append(messagesResponses.Messages, *response)
			}
		}

		response := new(Response)
		response.Data = messagesResponses

		if err := json.NewEncoder(w).Encode(response); err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func getStatusMessageHTTPHandler(ctx getMessageStatusContext) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")

		// get api key from header
		apiKey := r.Header.Get("X-Api-Key")

		// get user by api key
		user, err := ctx.UserStore.GetUserByAPIKey(apiKey)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// define id
		id, err := ulid.Parse(ps.ByName("id"))
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// get message by id
		msg, err := ctx.MessageStore.GetMessage(id)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get subject
		subject, err := ctx.SubjectStore.GeSubjectByID(msg.SubjectID)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate if message belongs to user
		if subject.UserID != user.ID {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// prepare response
		messageStatusResponse := new(MessageStatusResponse)
		messageStatusResponse.Status = string(msg.Status)

		response := new(Response)
		response.Data = messageStatusResponse

		if err := json.NewEncoder(w).Encode(response); err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func postCancelMessageHTTPHandler(ctx postCancelMessageContext) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")

		// Get api key from header
		apiKey := r.Header.Get("X-Api-Key")

		// Get user by api key
		user, err := ctx.UserStore.GetUserByAPIKey(apiKey)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Parse id
		id, err := ulid.Parse(ps.ByName("id"))
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get message
		msg, err := ctx.MessageStore.GetMessage(id)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get subject
		subject, err := ctx.SubjectStore.GeSubjectByID(msg.SubjectID)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate if message belongs to user
		if subject.UserID != user.ID {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// TODO(ca): check this
		addr := fmt.Sprintf("localhost:%d", grpcPort)

		// Grpc connection
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		// Define scheduler proto client
		client := proto.NewSchedulerServiceClient(conn)

		// delete message from scheduler
		_, err = client.Cancel(context.Background(), &proto.CancelRequest{
			Id: id.String(),
		})
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get message
		msg, err = ctx.MessageStore.GetMessage(id)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// prepare response
		messageCancelResponse := new(MessageCancelResponse)
		messageCancelResponse.Status = string(msg.Status)

		response := new(Response)
		response.Data = messageCancelResponse

		if err := json.NewEncoder(w).Encode(response); err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func httpLogginMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	log := logger.GetLogger().WithField("route", r.URL.Path)
	now := time.Now()

	ctx := context.WithValue(r.Context(), loggerKey, log)
	r = r.WithContext(ctx)

	log.Debug("Request received")
	next(w, r)
	log.WithField("elapsed", time.Since(now).String()).Debug("Request end")
}

func getLogger(r *http.Request) logger.Logger {
	return r.Context().Value(loggerKey).(logger.Logger)
}

func generateID(criteriaDelay time.Duration) (string, error) {
	delay := criteriaDelay

	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	id, err := ulid.New(
		ulid.Timestamp(time.Now().Add(delay)),
		entropy,
	)
	if err != nil {
		//TODO: move this message
		log.Println("Failed to create message id, %v", err)
		return "", err
	}

	return id.String(), nil
}

func sendMessage(client proto.SchedulerServiceClient, content string, endpoint string, subjectID string, criteriaDelay time.Duration) (string, error) {
	//generate uid
	id, err := generateID(criteriaDelay)
	if err != nil {
		return "", err
	}

	// put message to scheduler
	_, err = client.Put(context.Background(), &proto.PutRequest{
		Id:        id,
		Content:   []byte(content),
		Endpoint:  endpoint,
		SubjectId: subjectID,
	})
	if err != nil {
		// TODO: move this error
		log.Println("Put message failed, %v", err)
		return "", err
	}

	return id, nil
}

func getSubjectChannelByName(channelName string, subject *pigeon.Subject, cs *db.ChannelStore) (*pigeon.SubjectChannel, error) {
	var subjectChannel *pigeon.SubjectChannel

	// Set any subject channel instance
	for _, v := range subject.Channels {
		ch, err := cs.GetChannelById(v.ChannelID)
		if err != nil {
			continue
		}

		v.Channel = ch

		if ch.Name == channelName {
			subjectChannel = v
			break
		}
	}

	if subjectChannel == nil {
		return nil, errors.New(fmt.Sprintf("subject channel %s not found", channelName))
	}

	return subjectChannel, nil
}
