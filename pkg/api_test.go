package passport

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

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
				OpenFunc:        func() error { return nil },
				GetIdentityFunc: func(interaction Interaction) (*Identity, error) { return new(Identity), nil },
				CloseFunc:       func() error { return nil },
				CreateTableFunc: func() error { return nil },
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
				OpenFunc:        func() error { return nil },
				GetIdentityFunc: func(interaction Interaction) (*Identity, error) { return new(Identity), nil },
				CloseFunc:       func() error { return nil },
				CreateTableFunc: func() error { return nil },
			},
			ExpectedContentType: "application/json",
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
				OpenFunc:        func() error { return nil },
				GetIdentityFunc: func(interaction Interaction) (*Identity, error) { return new(Identity), nil },
				CloseFunc:       func() error { return nil },
				CreateTableFunc: func() error { return nil },
			},
			ExpectedContentType: "application/json",
		},
		// TODO| Test possible returned values for checkIdentity function
		// TODO| For an existent IP value => Returns an Identity struct => Identity {}
		// TODO| Create a flow diagram to represent the steps and the results that should return
		// TODO| and implement this logic in the tests.

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

func TestOpenDb(t *testing.T) {
	assert := assert.New(t)

	database := helperDb()

	err := database.Open()

	assert.NoError(err)
}

func TestCreateIdentity(t *testing.T) {
	assert := assert.New(t)

	database := helperDb()
	if err := database.Open(); err != nil {
		log.Fatalf("error opening database connection. err: %s", err)
	}
	defer database.Close()

	tests := []struct {
		Description string
		Database    Database
		Interaction Interaction
		Identity    Identity
		Idgroup     string
	}{
		{
			Description: "",
			Database:    database,
			Interaction: Interaction{
				IP:          "127.0.0.1",
				Provider:    "Test",
				Application: "Test",
			},
			Identity: Identity{},
			Idgroup:  "546dfa5sd4f6asd54f6as5d4f",
		},
		{
			Description: "",
			Database:    database,
			Interaction: Interaction{
				IP:          "127.0.0.1",
				Provider:    "Test",
				Application: "Test",
			},
			Identity: Identity{},
			Idgroup:  "",
		},
	}

	for _, test := range tests {
		test.Identity.createIdentity(test.Interaction, test.Idgroup)

		assert.Equal(test.Identity.Application, test.Interaction.Application)
		assert.Equal(test.Identity.IP, test.Interaction.IP)
		assert.Equal(test.Identity.Provider, test.Interaction.Provider)

		if test.Idgroup != "" {
			assert.Equal(test.Identity.Idgroup, test.Idgroup)
		}
	}
}

func TestCheckIdentity(t *testing.T) {
	assert := assert.New(t)

	database := helperDb()
	if err := database.Open(); err != nil {
		log.Fatalf("error opening database connection. err: %s", err)
	}
	defer database.Close()

	tests := []struct {
		Description string
		Database    Database
		Interaction Interaction
		Expectedout bool
		Idgroup     string
	}{
		{
			Description: "Case 1 IP value has coincidences",
			Database:    database,
			Interaction: Interaction{
				IP:          "127.0.0.1",
				Provider:    "Test Provider",
				Application: "Test Application",
			},
			Expectedout: false,
			Idgroup:     "221f168d-f97e-4ac5-910c-018b02d0ceb8",
		},
		{
			Description: "Case 2 IP value has not coincidences",
			Database:    database,
			Interaction: Interaction{
				IP:          helperRandstring(10),
				Provider:    "Test",
				Application: "Test",
			},
			Expectedout: true,
		},
	}

	for _, test := range tests {
		ident, out, err := database.checkIdentity(test.Interaction)

		assert.Equal(ident.IP, test.Interaction.IP)
		assert.Equal(ident.Provider, test.Interaction.Provider)
		assert.Equal(ident.Application, test.Interaction.Application)

		if out {
			assert.NotNil(ident.Idgroup)
			assert.NotNil(ident.Ididentity)
			assert.NotNil(ident.ID)
			database.db.Delete(ident)
		} else {
			assert.Equal(ident.Idgroup, test.Idgroup)
		}

		assert.Equal(out, test.Expectedout)
		assert.NoError(err)
	}
}

func TestCheckIdentitySecondLevel(t *testing.T) {
	assert := assert.New(t)

	database := helperDb()
	if err := database.Open(); err != nil {
		log.Fatalf("error opening database connection. err: %s", err)
	}
	defer database.Close()

	tests := []struct {
		Description string
		Database    Database
		Interaction Interaction
		Expectedout bool
		Idgroup     string
		ID          string
	}{
		{
			Description: "Case 1 is outside the time criteria [createdat < (now -2h)]",
			Database:    database,
			Interaction: Interaction{
				IP:          helperRandstring(10),
				Provider:    "Test Provider",
				Application: "Test Application",
			},
			Expectedout: false,
			Idgroup:     "221f168d-f97e-4ac5-910c-018b02d0ceb8",
		},
		{
			Description: "Case 2 is inside the time criteria [createdat < (now -2h)]",
			Database:    database,
			Interaction: Interaction{
				IP:          "8VFLYMXykR",
				Provider:    "Test",
				Application: "Test",
			},
			Expectedout: true,
			Idgroup:     "0988b32d-0781-42b1-a8da-f92d2b5e5c5a",
			ID:          "c894e0bc-7264-45df-9c81-50b0a694b357",
		},
	}

	for _, test := range tests {
		ident, err := database.checkIdentitySecondLevel(test.Interaction, test.Idgroup)

		assert.Equal(ident.IP, test.Interaction.IP)
		assert.Equal(ident.Application, test.Interaction.Application)
		assert.Equal(ident.Provider, test.Interaction.Provider)

		if test.Expectedout {
			assert.Equal(ident.ID, test.ID)
		} else {
			assert.Equal(ident.Idgroup, test.Idgroup)
			database.db.Delete(ident)
		}
		assert.NoError(err)
	}
}

func getSetting(setting string) string {
	value, ok := os.LookupEnv(setting)
	if !ok {
		log.Fatalf("Init error, %s ENV var not found", setting)
	}

	return value
}

func helperDb() Database {
	port := getSetting("DB_PORT")
	portInt, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		log.Fatalf("Error parsing to string Database's port %s, Err: %s", port, err)
	}

	database := Database{
		Host:      getSetting("DB_HOST"),
		Port:      portInt,
		User:      getSetting("DB_USER"),
		Password:  getSetting("DB_PASS"),
		DBName:    getSetting("DB_NAME"),
		Charset:   "utf8",
		ParseTime: "True",
		Loc:       "Local",
	}

	return database
}

func helperRandstring(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	seededRand := rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
