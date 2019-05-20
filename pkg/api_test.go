package idbsc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleFunction(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		Description         string
		Querier             Querier
		TypeRequest         string
		Interaction         Interaction
		StatusCode          int
		ExpectedContentType string
	}{
		{
			Description: "when HandleFunction receive a GET request",
			TypeRequest: http.MethodGet,
			StatusCode:  http.StatusMethodNotAllowed,
		},
		{
			Description: "when HandleFunction receive a POST request and an Interaction object",
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
				CloseFunc:         func() error { return nil },
				CreateTableFunc:   func() error { return nil },
			},
		},
		{
			Description: "when HandleFunction receive a POST request and an Interaction object and must receive an specific Content-Type header",
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
				CloseFunc:         func() error { return nil },
				CreateTableFunc:   func() error { return nil },
			},
			ExpectedContentType: "application/json",
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

		assert.Equal(test.StatusCode, resp.StatusCode)

		if (Interaction{}) != test.Interaction {
			assert.NotNil(resp)

			if len(test.ExpectedContentType) > 0 {
				assert.Equal(test.ExpectedContentType, resp.Header.Get("Content-Type"))
			}

			ident := new(Identity)
			if err := json.NewDecoder(resp.Body).Decode(ident); err != nil {
				t.Errorf("error unmarshaling the test response: Err: %v", err)
			}
			assert.NoError(err)
		}
	}
}
