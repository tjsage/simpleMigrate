package main

import (
	"flag"

	"github.com/labstack/gommon/log"
	"github.com/tjsage/simpleMigrate/migrate"
)

func main() {
	var dsn = flag.String("dsn", "root@tcp(localhost:3306)/TestDB", "MySQL DSN")
	var scriptsDirectory = flag.String("scripts", ".", "Path to scripts file")

	flag.Parse()

	err := migrate.Migrate(*dsn, *scriptsDirectory)
	if err != nil {
		log.Fatalf("program terminated early: %s", err.Error())
	}

}
