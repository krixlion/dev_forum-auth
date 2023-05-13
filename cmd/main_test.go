package main

import (
	"log"
	"testing"

	mongo "github.com/krixlion/dev_forum-auth/pkg/storage/mongo/testdata"
	vault "github.com/krixlion/dev_forum-auth/pkg/storage/vault/testdata"
	"go.uber.org/goleak"
)

// Avoid flags being parsed in the main's init func before they are defined by the testing init.
var _ = func() bool {
	testing.Init()
	return true
}()

func TestMain(m *testing.M) {
	if testing.Short() {
		goleak.VerifyTestMain(m)
	}

	if err := mongo.Seed(); err != nil {
		log.Fatalf("Failed to seed before testing: %v", err)
	}

	if err := vault.Seed(); err != nil {
		log.Fatalf("Failed to seed before testing: %v", err)
	}

	goleak.VerifyTestMain(m)
}
