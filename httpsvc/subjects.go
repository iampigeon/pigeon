package httpsvc

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/iampigeon/pigeon"
	"github.com/iampigeon/pigeon/db"
	"github.com/julienschmidt/httprouter"
)

// Contexts ...
type getSubjectsContext struct {
	SubjectStore *db.SubjectStore
	UserStore    *db.UserStore
	ChannelStore *db.ChannelStore
}

type postCreateSubjectContext struct {
	// MessageStore *db.MessageStore
	// SubjectStore *db.SubjectStore
	UserStore *db.UserStore
}

// Types ...

// SubjectChannel ...
// type SubjectChannel struct {
// 	ChannelID      string                 `json:"channel_id"`
// 	CriteriaID     string                 `json:"criteria_id"`
// 	CriteriaCustom float64                `json:"criteria_custom"`
// 	CallbackURL    string                 `json:"callback_post_url"`
// 	Options        map[string]interface{} `json:"options"`
// }

// Subject ...
// type Subject struct {
// 	Name     string            `json:"name"`
// 	Channels []*SubjectChannel `json:"channels"`
// }

// SubjectRequest ...
type SubjectRequest struct {
	Subject *pigeon.Subject `json:"subject"`
}

// SubjectResponse ...
type SubjectResponse struct {
	Name     string   `json:"name"`
	Channels []string `json:"channels"`
}

// SubjectsResponse ...
type SubjectsResponse struct {
	Subjects []*SubjectResponse `json:"subjects"`
}

// Handlers ...

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
		subjectsResponse.Subjects = make([]*SubjectResponse, 0)

		// get user subjects
		subjects, err := ctx.SubjectStore.GetSubjectsByUserID(user.ID)
		if err != nil {
			getLogger(r).Error(err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		for _, subject := range subjects {
			s := new(SubjectResponse)
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

func postCreateSubjectHTTPHandler(ctx postCreateSubjectContext) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		payload := new(SubjectRequest)

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

	}
}
