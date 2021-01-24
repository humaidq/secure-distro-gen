package models

import (
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite driver support
	"xorm.io/core"
	"xorm.io/xorm"
)

var (
	engine *xorm.Engine
	tables []interface{}
)

const DB_PATH = "./data.db"

func init() {
	tables = append(tables,
		new(Dummy),
	)
}

type Dummy struct {
}

// SetupEngine sets up an XORM engine according to the database configuration
// and syncs the schema.
func SetupEngine() *xorm.Engine {
	var err error
	engine, err = xorm.NewEngine("sqlite3", DB_PATH)

	if err != nil {
		log.Fatal("Unable to load the database! ", err)
	}

	engine.SetMapper(core.GonicMapper{}) // So ID becomes 'id' instead of 'i_d'
	err = engine.Sync(tables...)         // Sync the schema of tables

	if err != nil {
		log.Fatal("Unable to sync schema! ", err)
	}

	return engine
}
