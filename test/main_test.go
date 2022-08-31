package test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	db "github.com/junimslage10/gofinance-backend-auth/db/sqlc"
	env "github.com/junimslage10/gofinance-backend-auth/util"
	_ "github.com/lib/pq"
)

var testQueries *db.Queries

func TestMain(m *testing.M) {
	var dbDriver, filledStringDbSource, serverAddress = env.LoadEnv()
	conn, err := sql.Open(dbDriver, filledStringDbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	log.Print("connect to:", serverAddress)

	testQueries = db.New(conn)
	os.Exit(m.Run())
}
