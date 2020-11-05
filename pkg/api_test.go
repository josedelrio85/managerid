package managerid

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
		Description   string
		Interaction   Interaction
		Identity      Identity
		PassportIDGrp string
	}{
		{
			Description: "",
			Interaction: Interaction{
				IP:          "127.0.0.1",
				Provider:    "Test",
				Application: "Test",
			},
			Identity:      Identity{},
			PassportIDGrp: "546dfa5sd4f6asd54f6as5d4f",
		},
		{
			Description: "",
			Interaction: Interaction{
				IP:          "127.0.0.1",
				Provider:    "Test",
				Application: "Test",
			},
			Identity:      Identity{},
			PassportIDGrp: "",
		},
	}

	for _, test := range tests {
		test.Identity.createIdentity(test.Interaction, test.PassportIDGrp)

		assert.Equal(test.Identity.Application, test.Interaction.Application)
		assert.Equal(test.Identity.IP, test.Interaction.IP)
		assert.Equal(test.Identity.Provider, test.Interaction.Provider)

		if test.PassportIDGrp != "" {
			assert.Equal(test.Identity.PassportIDGrp, test.PassportIDGrp)
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
		Description   string
		Database      Database
		Interaction   Interaction
		Expectedout   bool
		PassportIDGrp string
		PassportID    string
	}{
		{
			Description: "Case 1 IP value has coincidences. Should not create any registry, reuse the matched result.",
			Interaction: Interaction{
				IP:          idt.IP,
				Provider:    idt.Provider,
				Application: idt.Application,
			},
			Expectedout:   false,
			PassportIDGrp: idt.PassportIDGrp,
			PassportID:    idt.PassportID,
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
			assert.NotNil(ident.PassportIDGrp)
			assert.NotNil(ident.Ididentity)
			assert.NotNil(ident.PassportID)
			assert.NotEqual(ident.PassportID, test.PassportID)
			identities = append(identities, *ident)
		} else {
			assert.Equal(ident.PassportIDGrp, test.PassportIDGrp)
			assert.Equal(ident.PassportID, test.PassportID)
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
		Description   string
		Interaction   Interaction
		Expectedout   bool
		PassportIDGrp string
		PassportID    string
	}{
		{
			Description: "Case 1 is outside the time criteria [createdat < (now -2h)]. Should create new registry with same values except ID value",
			Interaction: Interaction{
				IP:          ident1.IP,
				Provider:    ident1.Provider,
				Application: ident1.Application,
			},
			Expectedout:   false,
			PassportIDGrp: ident1.PassportIDGrp,
			PassportID:    ident1.PassportID,
		},
		{
			Description: "Case 2 is inside the time criteria [createdat < (now -2h)]. Should not create any registry, reuse the matched result. ID values must be equal",
			Interaction: Interaction{
				IP:          ident2.IP,
				Provider:    ident2.Provider,
				Application: ident2.Application,
			},
			Expectedout:   true,
			PassportIDGrp: ident2.PassportIDGrp,
			PassportID:    ident2.PassportID,
		},
	}

	for _, test := range tests {
		ident, err := dbInstance.checkIdentitySecondLevel(test.Interaction, test.PassportIDGrp)

		assert.Equal(ident.IP, test.Interaction.IP)
		assert.Equal(ident.Application, test.Interaction.Application)
		assert.Equal(ident.Provider, test.Interaction.Provider)

		if test.Expectedout {
			assert.Equal(ident.PassportID, test.PassportID)
		} else {
			assert.Equal(ident.PassportIDGrp, test.PassportIDGrp)
			assert.NotEqual(ident.PassportID, test.PassportID)
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

		if err := dbInstance.CreateTable(); err != nil {
			log.Printf("error creating the table. err: %s", err)
			return []Identity{}
		}

		ident := Identity{
			IP:            helperRandstring(10),
			Provider:      "TestProv",
			Application:   "TestApp",
			PassportIDGrp: fmt.Sprintf("%s", uuid.NewV4()),
			PassportID:    fmt.Sprintf("%s", uuid.NewV4()),
			Createdat:     hour,
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
