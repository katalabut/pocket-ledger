package main

import (
	"log"

	"github.com/katalabut/pocket-ledger/backend/internal/app"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	a, err := app.New()
	if err != nil {
		return err
	}

	a.Start()
	return nil
}
