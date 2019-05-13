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

	//It must be deleted when this first approach is closed
	f, err := os.OpenFile("../idbsc_log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	r := mux.NewRouter()

	host, ok := os.LookupEnv("TEST_HOST_IDBSC")
	if !ok {
		log.Fatal("Init error. Missing TEST_HOST_IDBSC ENV VAR")
	}

	port, ok := os.LookupEnv("TEST_PORT_IDBSC")
	if !ok {
		log.Fatal("Init error. Missing TEST_PORT_IDBSC ENV VAR")
	}

	user, ok := os.LookupEnv("TEST_USER_IDBSC")
	if !ok {
		log.Fatal("Init error. Missing TEST_USER_IDBSC ENV VAR")
	}

	password, ok := os.LookupEnv("TEST_PASSWORD_IDBSC")
	if !ok {
		log.Fatal("Init error. Missing TEST_PASSWORD_IDBSC ENV VAR")
	}

	dbname, ok := os.LookupEnv("TEST_DBNAME_IDBSC")
	if !ok {
		log.Fatal("Init error. Missing TEST_DBNAME_IDBSC ENV VAR")
	}

	portInt, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		log.Fatalf("Error parsing to string the RDS's port %s, Err: %s", port, err)
	}

	rds := &idbsc.Rdsgorm{
		Host:      host,
		Port:      portInt,
		User:      user,
		Password:  password,
		DBName:    dbname,
		Charset:   "utf8",
		ParseTime: "True",
		Loc:       "Local",
	}
	ch := idbsc.ClientHandler{
		Queriergorm: rds,
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
