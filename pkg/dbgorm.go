package idbsc

import (
	"fmt"
	"log"
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
}

// Identitygorm is a struct that represents an identity element.
type Identitygorm struct {
	IP          string `sql:"type:VARCHAR(255)"`
	Provider    string `sql:"type:VARCHAR(255)"`
	Application string `sql:"type:VARCHAR(255)"`
	Idgroup     string `sql:"type:VARCHAR(255)"`
	ID          string `sql:"type:VARCHAR(255)"`
	Createdat   *time.Time
	Ididentity  *int `gorm:"primary_key"`
}

func (Identitygorm) TableName() string {
	return "identities_bsc_gorm"
}

// Open opens a RDS connection using environment variable parameters.
// This implementation uses gorm library.
// Returns a db instance and nil if success or nil and Error instance if fails.
func (rg *Rdsgorm) Open() error {
	connstr := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=%v&parseTime=%v&loc=%v",
		rg.User, rg.Password, rg.Host, rg.Port, rg.DBName, rg.Charset, rg.ParseTime, rg.Loc)

	log.Println(connstr)

	db, err := gorm.Open("mysql", connstr)
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()

	if err = db.DB().Ping(); err != nil {
		log.Println(err)
		return err
	}

	rg.db = db

	return nil
}

// CreateTable blablabla
func (rg *Rdsgorm) CreateTable() error {
	rg.db.AutoMigrate(&Identitygorm{})

	log.Println("--CreateTable--")
	if !rg.db.HasTable(&Identitygorm{}) {
		log.Println("--No hay tabla, vamos a crearla--")
		rg.db.CreateTable(&Identitygorm{})
		log.Println("--Tabla creada!--")
	}
	return nil
}

// CheckIdentity queries for matches for the interaction struct passed as parameter.
// If no matches are returned, generates a new Identity using insert function.
// In case of there are some matches for the IP value, checks if complains the time criteria.
// In matches are returned, returns the identity element. In the other case, generates
// a new Bysidecar ID, and returns this identity element.
func (rg *Rdsgorm) CheckIdentity(interaction Interaction) (*Identitygorm, error) {
	// sqlQ := fmt.Sprintf(`select ip, application, provider, idgroup, id, createdat
	// from identities_bsc where ip = '%s' order by createdat desc limit 1`,
	// 	interaction.IP)

	fmt.Println(interaction)
	// ident := new(Identitygorm)
	ident := make([]Identitygorm, 0)
	rg.db.Debug().Where("ip = ?", interaction.IP).Order("createddat desc").First(&ident)
	// rg.db.Debug().Table("identities_bsc").Find(&ident)

	log.Println(ident)
	return nil, nil
}
