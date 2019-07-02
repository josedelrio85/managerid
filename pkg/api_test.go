package passport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

var dbInstance Database
var identities []Identity

func TestMain(m *testing.M) {
	dbInstance = helperDb()
	identities = populateDb(2)

	code := m.Run()

	setDownDb()

	os.Exit(code)
}

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

	err := dbInstance.Open()

	assert.NoError(err)
}

func TestCreateIdentity(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		Description string
		Interaction Interaction
		Identity    Identity
		Idgroup     string
	}{
		{
			Description: "",
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

	if err := dbInstance.Open(); err != nil {
		t.Errorf("error opening database connection. err: %s", err)
	}

	var idt Identity
	for i, z := range identities {
		if i%2 != 0 {
			idt = z
		}
	}

	tests := []struct {
		Description string
		Database    Database
		Interaction Interaction
		Expectedout bool
		Idgroup     string
		ID          string
	}{
		{
			Description: "Case 1 IP value has coincidences. Should not create any registry, reuse the matched result.",
			Interaction: Interaction{
				IP:          idt.IP,
				Provider:    idt.Provider,
				Application: idt.Application,
			},
			Expectedout: false,
			Idgroup:     idt.Idgroup,
			ID:          idt.ID,
		},
		{
			Description: "Case 2 IP value has not coincidences. Should create new registry.",
			Interaction: Interaction{
				IP:          helperRandstring(10),
				Provider:    "Test",
				Application: "Test",
			},
			Expectedout: true,
		},
	}

	for _, test := range tests {
		ident, out, err := dbInstance.checkIdentity(test.Interaction)

		assert.Equal(ident.IP, test.Interaction.IP)
		assert.Equal(ident.Provider, test.Interaction.Provider)
		assert.Equal(ident.Application, test.Interaction.Application)

		if out {
			assert.NotNil(ident.Idgroup)
			assert.NotNil(ident.Ididentity)
			assert.NotNil(ident.ID)
			assert.NotEqual(ident.ID, test.ID)
			identities = append(identities, *ident)
		} else {
			assert.Equal(ident.Idgroup, test.Idgroup)
			assert.Equal(ident.ID, test.ID)
		}
		assert.Equal(out, test.Expectedout)
		assert.NoError(err)
	}
}

func TestCheckIdentitySecondLevel(t *testing.T) {
	assert := assert.New(t)

	if err := dbInstance.Open(); err != nil {
		t.Errorf("error opening database connection. err: %s", err)
	}

	var ident1 Identity
	var ident2 Identity
	for i, z := range identities {
		if i%2 == 0 {
			ident2 = z
		} else {
			ident1 = z
		}
	}

	tests := []struct {
		Description string
		Interaction Interaction
		Expectedout bool
		Idgroup     string
		ID          string
	}{
		{
			Description: "Case 1 is outside the time criteria [createdat < (now -2h)]. Should create new registry with same values except ID value",
			Interaction: Interaction{
				IP:          ident1.IP,
				Provider:    ident1.Provider,
				Application: ident1.Application,
			},
			Expectedout: false,
			Idgroup:     ident1.Idgroup,
			ID:          ident1.ID,
		},
		{
			Description: "Case 2 is inside the time criteria [createdat < (now -2h)]. Should not create any registry, reuse the matched result. ID values must be equal",
			Interaction: Interaction{
				IP:          ident2.IP,
				Provider:    ident2.Provider,
				Application: ident2.Application,
			},
			Expectedout: true,
			Idgroup:     ident2.Idgroup,
			ID:          ident2.ID,
		},
	}

	for _, test := range tests {
		ident, err := dbInstance.checkIdentitySecondLevel(test.Interaction, test.Idgroup)

		assert.Equal(ident.IP, test.Interaction.IP)
		assert.Equal(ident.Application, test.Interaction.Application)
		assert.Equal(ident.Provider, test.Interaction.Provider)

		if test.Expectedout {
			assert.Equal(ident.ID, test.ID)
		} else {
			assert.Equal(ident.Idgroup, test.Idgroup)
			assert.NotEqual(ident.ID, test.ID)
			// dbInstance.db.Delete(&ident)
			identities = append(identities, *ident)
		}
		assert.NoError(err)
	}
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

func populateDb(number int) []Identity {
	if err := dbInstance.Open(); err != nil {
		log.Printf("error opening database connection. err: %s", err)
		return []Identity{}
	}
	defer dbInstance.Close()

	for i := 1; i <= number; i++ {
		hour := time.Now()
		if i%2 == 0 {
			hour = time.Now().Add(time.Duration(-150) * time.Minute)
		}

		ident := Identity{
			IP:          helperRandstring(10),
			Provider:    "TestProv",
			Application: "TestApp",
			Idgroup:     fmt.Sprintf("%s", uuid.NewV4()),
			ID:          fmt.Sprintf("%s", uuid.NewV4()),
			Createdat:   hour,
		}

		dbInstance.db.Create(&ident)
		identities = append(identities, ident)
	}
	return identities
}

func setDownDb() {

	if err := dbInstance.Open(); err != nil {
		log.Printf("error opening database connection. err: %s", err)
	}
	defer dbInstance.Close()

	for _, ident := range identities {
		dbInstance.db.Delete(&ident)
	}
}

func getSetting(setting string) string {
	value, ok := os.LookupEnv(setting)
	if !ok {
		log.Fatalf("Init error, %s ENV var not found", setting)
	}

	return value
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
