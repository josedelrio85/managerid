package passport

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // go mysql driver
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // mysql import driver for gorm
	uuid "github.com/satori/go.uuid"
)

// Database is a struct to manage DB environment configuration.
type Database struct {
	Host      string
	Port      int64
	User      string
	Password  string
	DBName    string
	Charset   string
	ParseTime string
	Loc       string

	db *gorm.DB
}

// Querier is an interface used to force client handler to implement
// Open, GetIdentity, Close and CreateTable methods
type Querier interface {
	Open() error
	GetIdentity(Interaction) (*Identity, error)
	Close()
	CreateTable() error
}

// Identity is a struct that represents an identity element.
type Identity struct {
	IP          string    `sql:"type:VARCHAR(255)" json:"-"`
	Provider    string    `sql:"type:VARCHAR(255)" json:"-"`
	Application string    `sql:"type:VARCHAR(255)" json:"-"`
	Idgroup     string    `sql:"type:VARCHAR(255)" json:"idgroup"`
	ID          string    `sql:"type:VARCHAR(255)" json:"idbysidecar"`
	Createdat   time.Time `json:"-"`
	Ididentity  *int      `gorm:"primary_key" json:"-"`
}

// TableName sets the default table name
func (Identity) TableName() string {
	return "identities_bsc"
}

// Open opens a RDS connection using environment variable parameters.
// This implementation uses gorm library.
// Returns a db instance and nil if success or nil and Error instance if fails.
func (rg *Database) Open() error {
	connstr := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=%v&parseTime=%v&loc=%v",
		rg.User, rg.Password, rg.Host, rg.Port, rg.DBName, rg.Charset, rg.ParseTime, rg.Loc)

	db, err := gorm.Open("mysql", connstr)
	if err != nil {
		return err
	}

	if err = db.DB().Ping(); err != nil {
		return err
	}

	rg.db = db

	return nil
}

// Close Database.db instance
func (rg *Database) Close() {
	rg.db.Close()
}

// CreateTable blablabla
func (rg *Database) CreateTable() error {
	rg.db.AutoMigrate(&Identity{})

	if !rg.db.HasTable(&Identity{}) {
		rg.db.CreateTable(&Identity{})
	}
	return nil
}

// GetIdentity queries for matches for the interaction struct passed as parameter.
// If no matches are returned, generates a new Identity using insert function.
// In case of there are some matches for the IP value, checks if complains the time criteria.
// If matches are returned, returns the identity element.
// In other case, generates a new Bysidecar ID, and returns this identity element.
func (rg *Database) GetIdentity(interaction Interaction) (*Identity, error) {

	// check if there are no results =>  create idgroup and id and store in DB
	ident, out, err := rg.checkIdentity(interaction)
	if err != nil {
		return nil, err
	}

	if out {
		return ident, nil
	}

	ident, err = rg.checkIdentitySecondLevel(interaction, ident.Idgroup)
	if err != nil {
		return nil, err
	}

	// returns the ident resultant object
	return ident, nil
}

// checkIdentity checks if there is any row that matches the IP criteria.
// if there are no results, creates a new identity and set out is true to end the execution.
// returns the identity true|false nil || nil false error
func (rg *Database) checkIdentity(interaction Interaction) (*Identity, bool, error) {
	ident := new(Identity)
	out := false
	err := rg.db.Where("ip = ?", interaction.IP).Order("createdat desc").First(&ident).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, out, err
	}

	if gorm.IsRecordNotFoundError(err) {
		ident.createIdentity(interaction, "")
		rg.db.Create(ident)
		out = true
	}
	return ident, out, nil
}

// checkIdentitySecondLevel checks if there is any row that matches the
// IP+Applicaciont+Provider+TwoHourless criteria.
// returns a pointer to the matched identity
func (rg *Database) checkIdentitySecondLevel(interaction Interaction, idgroup string) (*Identity, error) {
	ident := new(Identity)
	twoHoursLess := time.Now().Add(time.Duration(-120) * time.Minute)
	timeFormatted := twoHoursLess.Format("2006-01-02 15:04:05")

	err := rg.db.Where("ip = ? and provider = ? and application = ? and createdat > ?",
		interaction.IP,
		interaction.Provider,
		interaction.Application,
		timeFormatted).Order("createdat desc").First(&ident).Error

	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	if gorm.IsRecordNotFoundError(err) {
		ident.createIdentity(interaction, idgroup)
		rg.db.Create(ident)
	}

	return ident, nil
}

// createIdentity creates an Identity object with proper values
// If idgroup is setted you must use as Idgroup value
func (ident *Identity) createIdentity(interaction Interaction, idgroup string) {

	ident.Application = interaction.Application
	ident.IP = interaction.IP
	ident.Provider = interaction.Provider
	ident.Createdat = time.Now()
	ident.ID = fmt.Sprintf("%s", uuid.NewV4())
	ident.Idgroup = fmt.Sprintf("%s", uuid.NewV4())

	if idgroup != "" {
		ident.Ididentity = nil
		ident.Idgroup = idgroup
	}
}
