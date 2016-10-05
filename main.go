package main

import (
	"flag"
	"os"

	"github.com/labstack/gommon/log"
	"github.com/tjsage/simpleMigrate/migrate"
)

func main() {
	var dsn = flag.String("dsn", os.Getenv("SIMPLE_MIGRATE_DSN"), "MySQL DSN")
	var scriptsDirectory = flag.String("scripts", ".", "Path to scripts file")

	flag.Parse()

	err := migrate.Migrate(*dsn, *scriptsDirectory)
	if err != nil {
		log.Fatalf("program terminated early: %s", err.Error())
	}

}
