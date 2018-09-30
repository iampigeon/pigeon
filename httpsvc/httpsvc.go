package httpsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/WiseGrowth/go-wisebot/logger"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

type key int

const (
	httpPort      = 5252
	loggerKey key = iota

	APIKey = "12345"
)

// {
// "data": {
// 	"subjects": [{
// 	"name": "max-air-temperature",
// 	"channels": ["mqtt", "sms"]
// 	}, {
// 	"name": "weekly-report",
// 	"channels": ["sms", "push"]
// 	}]
// },
// "meta": {}
// }

// Subject ...
type Subject struct {
	Name     string   `json:"name"`
	Channels []string `json:"channels"`
}

// SubjectsResponse ...
type SubjectsResponse struct {
	Subjects []Subject `json:"subjects"`
}

// Response ...
type Response struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta"`
}

// NewHTTPServer returns an initialized server
//
// GET /subjects
//
func NewHTTPServer() *http.Server {
	router := httprouter.New()

	router.GET("/api/v1/subjects", subjectsHTTPHandler)

	addr := fmt.Sprintf(":%d", httpPort)
	routes := negroni.Wrap(router)
	n := negroni.New(negroni.HandlerFunc(httpLogginMiddleware), routes)
	server := &http.Server{Addr: addr, Handler: n}

	return server
}

func subjectsHTTPHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	// TODO: implement logic to validate request by x-api-key header value
	apiKey := r.Header.Get("X-Api-Key")
	if apiKey != APIKey {
		err := errors.New("invalid X-Api-Key header param")
		getLogger(r).Error(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	//TODO: implements this
	s := new(Subject)
	s.Name = "max-air-temperature"
	s.Channels = append(s.Channels, "mqtt", "sms")
	res := new(SubjectsResponse)
	res.Subjects = append(res.Subjects, *s)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		getLogger(r).Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
