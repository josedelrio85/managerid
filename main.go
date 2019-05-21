package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	idbsc "github.com/bysidecar/idbsc/pkg"
	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()

	port, ok := os.LookupEnv("PORT_IDBSC")
	if !ok {
		log.Fatal("Init error. Missing PORT_IDBSC ENV VAR")
	}

	portInt, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		log.Fatalf("Error parsing to string the RDS's port %s, Err: %s", port, err)
	}

	rds := &idbsc.Database{
		Host:      GetSetting("HOST_IDBSC"),
		Port:      portInt,
		User:      GetSetting("USER_IDBSC"),
		Password:  GetSetting("PASSWORD_IDBSC"),
		DBName:    GetSetting("DBNAME_IDBSC"),
		Charset:   "utf8",
		ParseTime: "True",
		Loc:       "Local",
	}
	ch := idbsc.ClientHandler{
		Querier: rds,
	}

	if err := rds.Open(); err != nil {
		log.Fatalf("error opening redshift's connection. err: %s", err)
	}
	defer rds.Close()

	if err := rds.CreateTable(); err != nil {
		log.Fatalf("error creating the table. err: %s", err)
	}

	r.PathPrefix("/idbsc/").Handler(ch.HandleFunction())

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
