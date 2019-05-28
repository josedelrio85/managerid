package passport

import (
	"log"
	"encoding/json"
	"fmt"
	"net/http"
)

// Interaction is a struct that represents a single interaction in web environment.
type Interaction struct {
	IP          string `json:"ip"`
	Provider    string `json:"provider"`
	Application string `json:"application"`
}

// ClientHandler is a struct created to use its ch property as element that implements
// http.Handler.Neededed to call HandleFunction as param in router Handler function.
type ClientHandler struct {
	ch      http.Handler
	Interac Interaction
	Querier Querier
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
			log.Println("Method not allowed.", http.StatusMethodNotAllowed)
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&ch.Interac); err != nil {
			message := fmt.Sprintf("error decoding interaction payload, err: %v", err)
			log.Println(message)
			http.Error(w, message, http.StatusInternalServerError)
			return
		}

		identity, err := ch.Querier.GetIdentity(ch.Interac)
		if err != nil {
			message := fmt.Sprintf("error performing interaction's CheckIdentity, err: %v", err)
			log.Println(message)
			http.Error(w, message, http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(identity)
	})
}
