package migrate

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db                 *sql.DB
	nonEmptyQueryRegex = regexp.MustCompile(`\w+`)
)

func Migrate(dsn string, scriptDirectory string) error {
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("unable to ping database: %s", err.Error())
	}

	err = createMigrationTableIfNotExists()
	if err != nil {
		return fmt.Errorf("failed to create table: %s", err.Error())
	}

	ranScripts, err := getRanScripts()
	if err != nil {
		return fmt.Errorf("unable to get ran scripts: %s", err.Error())
	}

	files, err := getScriptFiles(scriptDirectory)
	if err != nil {
		return fmt.Errorf("unable to find scripts: %s", err.Error())
	}

	err = runNewMigrationScripts(ranScripts, scriptDirectory, files)
	if err != nil {
		return fmt.Errorf("unable to run all scripts: %s", err.Error())
	}

	return nil
}

func createMigrationTableIfNotExists() error {
	rows, err := db.Query("SHOW TABLES LIKE 'migrations'")
	if err != nil {
		return err
	}
	defer rows.Close()

	// Table doesn't exist
	if !rows.Next() {
		_, err := db.Exec(createMigrationTableSQL)
		if err != nil {
			return err
		}
	}

	return nil
}

func getRanScripts() (scripts map[string]bool, err error) {
	scripts = make(map[string]bool)

	rows, err := db.Query(getRanMigrationsSQL)
	if err != nil {
		return scripts, err
	}

	for rows.Next() {
		var scriptName string
		err = rows.Scan(&scriptName)
		if err != nil {
			return scripts, err
		}

		scripts[scriptName] = true
	}

	return scripts, nil
}

func getScriptFiles(directory string) (scripts []string, err error) {
	files, err := ioutil.ReadDir(directory)

	if err != nil {
		return scripts, err
	}

	for _, file := range files {
		fmt.Println(file.Name())
		if !file.IsDir() {
			scripts = append(scripts, file.Name())
		}
	}

	return scripts, err
}

func runNewMigrationScripts(ranScripts map[string]bool, directory string, files []string) error {
	for _, script := range files {
		if ranScripts[script] {
			log.Printf("Skipping: %s, already ran", script)
			continue
		}

		log.Printf("Running: %s", script)
		err := runScript(filepath.Join(directory, script))
		if err != nil {
			return err
		}

		err = recordScriptRun(script)
		if err != nil {
			return err
		}
	}

	return nil
}

func recordScriptRun(file string) error {
	_, err := db.Exec("INSERT INTO migrations (script, date_ran) VALUES (?, ?)", file, time.Now())
	return err
}

func runScript(path string) error {
	log.Printf("Reading file: %s", path)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("unable to read file %s, error: %s", path, err.Error())
	}

	scripts := strings.Split(string(data), ";")

	for _, script := range scripts {
		// Make sure script isn't empty.
		empty := !nonEmptyQueryRegex.MatchString(script)
		if empty {
			continue
		}

		_, err = db.Exec(script)
		if err != nil {
			return fmt.Errorf("uanble to execute script %s, error: %s", path, err.Error())
		}
	}

	return nil
}

var createMigrationTableSQL = `
	CREATE TABLE migrations (
		script VARCHAR(255) NOT NULL,
		date_ran DATETIME NOT NULL
	);
`

var getRanMigrationsSQL = `
	SELECT
		script
	FROM migrations
`
