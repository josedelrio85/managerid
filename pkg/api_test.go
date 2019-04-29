package idbsc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleFunction(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		Description string
		Querier     Querier
		TypeRequest string
		Interaction Interaction
		StatusCode  int
	}{
		// {
		// 	Description: "when HandleFunction receive a GET request",
		// 	TypeRequest: http.MethodGet,
		// 	StatusCode:  http.StatusAccepted,
		// },
		{
			Description: "when HandleFunction receive a POST request",
			TypeRequest: http.MethodPost,
			StatusCode:  http.StatusOK,
			Interaction: Interaction{
				IP:          "127.0.0.1",
				Application: "Test Application",
				Provider:    "Test Provider",
			},
			Querier: &FakeDb{
				OpenFunc:          func() error { return nil },
				CheckIdentityFunc: func(interaction Interaction) (*Identity, error) { return new(Identity), nil },
			},
		},
	}

	for _, test := range tests {
		ch := ClientHandler{
			Querier: test.Querier,
		}
		ts := httptest.NewServer(ch.HandleFunction())
		defer ts.Close()

		body, err := json.Marshal(test.Interaction)
		if err != nil {
			t.Errorf("error marshaling test json: Err: %v", err)
			return
		}

		req, err := http.NewRequest(test.TypeRequest, ts.URL, bytes.NewBuffer(body))
		if err != nil {
			t.Errorf("error creating the test Request: Err: %v", err)
			return
		}

		http := &http.Client{}
		resp, err := http.Do(req)
		if err != nil {
			t.Errorf("error sending test Request: Err: %v", err)
			return
		}
		fmt.Println(resp)

		assert.Equal(test.StatusCode, 200)
	}

	// 	assert.Equal(test.StatusCode, resp.StatusCode)
	// }
}
