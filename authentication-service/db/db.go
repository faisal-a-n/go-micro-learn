package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var count int64

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func ConnectToDB() *sql.DB {
	dsn := os.Getenv("DB_SOURCE")
	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not ready yet")
			count++
		} else {
			log.Println("Connected to Postgres")
			return connection
		}
		if count > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Waiting for 2 seconds")
		time.Sleep(time.Second * 2)
		continue
	}
}
