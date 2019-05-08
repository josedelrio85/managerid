package idbsc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Interaction is a struct that represents a single interaction in web environment.
type Interaction struct {
	IP          string `json:"ip"`
	Provider    string `json:"provider"`
	Application string `json:"application"`
}

// Identity is a struct that represents an identity element.
type Identity struct {
	IP          string
	Provider    string
	Application string
	Idgroup     string
	ID          string
	Createdat   time.Time
	Ididentity  int
}

// ClientHandler is a struct created to use its ch property as element that implements
// http.Handler.Neededed to call HandleFunction as param in router Handler function.
type ClientHandler struct {
	ch          http.Handler
	Interac     Interaction
	Querier     Querier
	Queriergorm Queriergorm
}

// HandleFunction is a function used to manage all received requests.
// Only POST method accepted.
// Decode the identity json request as Identity struct.
// Check if the data has matches in DB environment to make a decission.
// Returns an StatusMethodNotAllowed state if other kind of request is received.
// Returns StatusInternalServerError when decoding the body content fails.
func (ch *ClientHandler) HandleFunction() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&ch.Interac); err != nil {
			message := fmt.Sprintf("error decoding interaction payload, err: %v", err)
			http.Error(w, message, http.StatusInternalServerError)
			return
		}

		// identity, err := ch.Querier.CheckIdentity(ch.Interac)
		identity, err := ch.Queriergorm.CheckIdentity(ch.Interac)
		if err != nil {
			message := fmt.Sprintf("error performing interaction's CheckIdentity, err: %v", err)
			http.Error(w, message, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(identity)
	})
}
