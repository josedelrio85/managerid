package idbsc

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // go mysql driver
)

// Rds is a struct to manage AWS RDS's environment configuration.
type Rds struct {
	Host     string
	Port     int64
	User     string
	Password string
	DBName   string

	db *sql.DB
}

// Querier is an interface used to force client handler to implement
// Open and CheckIdentity methods
type Querier interface {
	Open() error
	CheckIdentity(Interaction) (*Identity, error)
}

// Open opens a RDS connection using environment variable parameters.
// Returns a db instance and nil if success or nil and Error instance if fails.
func (r *Rds) Open() error {
	connstring := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v",
		r.User, r.Password, r.Host, r.Port, r.DBName)

	db, err := sql.Open("mysql", connstring)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	r.db = db
	return nil
}

// CheckIdentity queries for matches for the interaction struct passed as parameter.
// If no matches are returned, generates a new Identity using insert function.
// In case of there are some matches for the IP value, checks if complains the time criteria.
// In matches are returned, returns the identity element. In the other case, generates
// a new Bysidecar ID, and returns this identity element.
func (r *Rds) CheckIdentity(interaction Interaction) (*Identity, error) {

	sqlQ := fmt.Sprintf(`select ip, application, provider, idgroup, id, createdat 
	from identities_bsc where ip = '%s' order by createdat desc limit 1`,
		interaction.IP)
	row := r.db.QueryRow(sqlQ)

	ident := new(Identity)
	var timetest = ""
	if err := row.Scan(&ident.IP, &ident.Application, &ident.Provider, &ident.Idgroup, &ident.ID, &timetest); err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if err == sql.ErrNoRows {
		// no results, generate idgroup and id
		ident.Application = interaction.Application
		ident.IP = interaction.IP
		ident.Provider = interaction.Provider

		uuidgroup, err := getUUID()
		if err != nil {
			return nil, err
		}
		ident.Idgroup = fmt.Sprintf("%s", uuidgroup)

		uuid, err := getUUID()
		if err != nil {
			return nil, err
		}
		ident.ID = fmt.Sprintf("%s", uuid)

		r.insert(ident)
		return ident, nil
	}

	// there are at least one match for the IP value, check other conditions
	twoHoursLess := time.Now().Add(time.Duration(-120) * time.Minute)

	sqlLevelTwo := fmt.Sprintf(`select ip, provider, application, idgroup, id
	from identities_bsc where ip= '%s' and provider = '%s' and application= '%s' and createdat > '%s' order by createdat desc limit 1`,
		interaction.IP, interaction.Provider, interaction.Application, twoHoursLess.Format("2006-01-02 15:04:05"))

	rr := r.db.QueryRow(sqlLevelTwo)
	if nexterr := rr.Scan(&ident.IP, &ident.Provider, &ident.Application, &ident.Idgroup, &ident.ID); nexterr != nil && nexterr != sql.ErrNoRows { //
		return nil, nexterr
	} else if nexterr == sql.ErrNoRows {
		// generate ID and reuse idgroup
		uuid, err := getUUID()
		if err != nil {
			return nil, err
		}
		ident.ID = fmt.Sprintf("%s", uuid)

		r.insert(ident)
	}
	// returns the last row that matches the criteria (neither new id or idgroup is created)
	return ident, nil
}

// insert generates a new row in identities_bsc table.
// Returns true and nil error if success.
// Returns false and error element if failed.
func (r *Rds) insert(identity *Identity) (bool, error) {
	sqlInsert := `insert into identities_bsc (ip, provider, application, idgroup, id, createdat) 
	values (?,?,?,?,?,?);`

	stmt, _ := r.db.Prepare(sqlInsert)
	_, err := stmt.Exec(identity.IP, identity.Provider, identity.Application, identity.Idgroup, identity.ID, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return false, err
	}
	return true, nil
}
