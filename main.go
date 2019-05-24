package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	passport "github.com/bysidecar/passport/pkg"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	port := GetSetting("DB_PORT")
	portInt, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		log.Fatalf("Error parsing to string Database's port %s, Err: %s", port, err)
	}

	database := &passport.Database{
		Host:      GetSetting("DB_HOST"),
		Port:      portInt,
		User:      GetSetting("DB_USER"),
		Password:  GetSetting("DB_PASS"),
		DBName:    GetSetting("DB_NAME"),
		Charset:   "utf8",
		ParseTime: "True",
		Loc:       "Local",
	}
	ch := passport.ClientHandler{
		Querier: database,
	}

	if err := database.Open(); err != nil {
		log.Fatalf("error opening database connection. err: %s", err)
	}
	defer database.Close()

	if err := database.CreateTable(); err != nil {
		log.Fatalf("error creating the table. err: %s", err)
	}

	r.PathPrefix("/id/settle").Handler(ch.HandleFunction())

	log.Fatal(http.ListenAndServe(":4000", r))
}

// GetSetting reads an ENV VAR setting, it does crash the service if with an
// error message if any setting is not found.
//
// - setting: The setting (ENV VAR) to read.
//
// Returns the setting value.
func GetSetting(setting string) string {
	value, ok := os.LookupEnv(setting)
	if !ok {
		log.Fatalf("Init error, %s ENV var not found", setting)
	}

	return value
}
