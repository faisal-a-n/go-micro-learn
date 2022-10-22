package main

import (
	"authentication-service/data"
	database "authentication-service/db"
	"database/sql"
	"log"
	"net/http"
)

const PORT = "0.0.0.0:80"

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Authentication service running")

	//Connect to DB
	conn := database.ConnectToDB()

	if conn == nil {
		log.Println("Couldn't connect to postgres")
	}

	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	server := http.Server{
		Addr:    PORT,
		Handler: app.routes(),
	}

	err := server.ListenAndServe()

	if err != nil {
		log.Fatal("there was an error starting authentication-service server", err)
	}

}
