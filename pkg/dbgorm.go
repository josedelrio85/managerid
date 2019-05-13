package idbsc

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // go mysql driver
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // mysql import driver for gorm
)

// Rdsgorm is a struct to manage AWS RDS's environment configuration.
type Rdsgorm struct {
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

// Queriergorm is an interface used to force client handler to implement
// Open and CheckIdentity methods
type Queriergorm interface {
	Open() error
	CheckIdentity(Interaction) (*Identitygorm, error)
	Close()
	CreateTable() error
}

// Identitygorm is a struct that represents an identity element.
type Identitygorm struct {
	IP          string `sql:"type:VARCHAR(255)"`
	Provider    string `sql:"type:VARCHAR(255)"`
	Application string `sql:"type:VARCHAR(255)"`
	Idgroup     string `sql:"type:VARCHAR(255)"`
	ID          string `sql:"type:VARCHAR(255)"`
	Createdat   time.Time
	Ididentity  *int `gorm:"primary_key"`
}

// TableName sets the default table name
func (Identitygorm) TableName() string {
	return "identities_bsc"
}

// Open opens a RDS connection using environment variable parameters.
// This implementation uses gorm library.
// Returns a db instance and nil if success or nil and Error instance if fails.
func (rg *Rdsgorm) Open() error {
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

// Close Rdsgorm.db instance
func (rg *Rdsgorm) Close() {
	rg.db.Close()
}

// CreateTable blablabla
func (rg *Rdsgorm) CreateTable() error {
	rg.db.AutoMigrate(&Identitygorm{})

	if !rg.db.HasTable(&Identitygorm{}) {
		rg.db.CreateTable(&Identitygorm{})
	}
	return nil
}

// CheckIdentity queries for matches for the interaction struct passed as parameter.
// If no matches are returned, generates a new Identity using insert function.
// In case of there are some matches for the IP value, checks if complains the time criteria.
// In matches are returned, returns the identity element. In the other case, generates
// a new Bysidecar ID, and returns this identity element.
func (rg *Rdsgorm) CheckIdentity(interaction Interaction) (*Identitygorm, error) {
	ident := new(Identitygorm)
	err := rg.db.Where("ip = ?", interaction.IP).Order("createdat desc").First(&ident).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	} else if gorm.IsRecordNotFoundError(err) {
		// no results, create idgroup and id
		if err := ident.createIdentity(interaction); err != nil {
			return nil, err
		}
		rg.db.Create(ident)
		return ident, nil
	}

	// there are at least one match for the IP value, check other conditions
	twoHoursLess := time.Now().Add(time.Duration(-120) * time.Minute)

	err = rg.db.Where("ip = ? and provider = ? and application = ? and createdat > ?",
		interaction.IP, interaction.Provider, interaction.Application, twoHoursLess.Format("2006-01-02 15:04:05")).First(&ident).Error

	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	} else if gorm.IsRecordNotFoundError(err) {
		// generate ID and reuse idgroup
		uuid, err := getUUID()
		if err != nil {
			return nil, err
		}
		ident.ID = fmt.Sprintf("%s", uuid)
		ident.Ididentity = nil
		ident.Createdat = time.Now()

		rg.db.Create(ident)
		return ident, nil
	}
	// returns the last row that matches the criteria (neither new id or idgroup is created)
	return ident, nil
}

func (ident *Identitygorm) createIdentity(interaction Interaction) error {

	uuidgroup, err := getUUID()
	if err != nil {
		return err
	}
	uuid, err := getUUID()
	if err != nil {
		return err
	}

	ident.Application = interaction.Application
	ident.IP = interaction.IP
	ident.Provider = interaction.Provider
	ident.Createdat = time.Now()
	ident.ID = fmt.Sprintf("%s", uuid)
	ident.Idgroup = fmt.Sprintf("%s", uuidgroup)

	return nil
}
