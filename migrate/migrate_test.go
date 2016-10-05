package migrate

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestGetFiles(t *testing.T) {
	files, err := getScriptFiles("../testScripts")

	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(files))
	assert.Equal(t, "001_init.sql", files[0])
}

func TestMigration(t *testing.T) {
	dsn := os.Getenv("SIMPLE_MIGRATE_DSN")
	// defer cleanup(dsn)

	// Run migration process once.
	err := Migrate(dsn, "../testScripts")
	assert.Equal(t, nil, err)

	// Add another test file, run through it again.
	err = ioutil.WriteFile("../testScripts/002_ADD_CAT.sql", []byte(testScript), 777)
	assert.Equal(t, nil, err)

	// Run second time
	err = Migrate(dsn, "../testScripts")
	assert.Equal(t, nil, err)

}

func cleanup(dsn string) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("[ERROR] Error opening db connection: %s", err.Error())
	}
	defer db.Close()

	_, err = db.Exec("DROP TABLE DUCKS")
	if err != nil {
		log.Printf("[ERROR] Failed to drop test table DUCKS: %s", err.Error())
	}

	_, err = db.Exec("DROP TABLE CATS")
	if err != nil {
		log.Printf("[ERROR] Failed to drop table CATS: %s", err.Error())
	}

	err = os.Remove("../testScripts/002_ADD_CAT.sql")
	if err != nil {
		log.Printf("[ERROR] FAiled to remove test script 2: %s", err.Error())
	}

}

var testScript string = `
	CREATE TABLE Cats (
		CatID INT PRIMARY KEY NOT NULL AUTO_INCREMENT,
		Name VARCHAR(255) NOT NULL
	);
`
