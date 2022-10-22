package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

const PORT = "0.0.0.0:80"

type Config struct {
	Mailer Mail
}

func main() {
	app := Config{
		Mailer: newMail(),
	}
	log.Println("Starting mail service on ", PORT)

	server := &http.Server{
		Addr:    PORT,
		Handler: app.routes(),
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Panic("Cannot start mail service ", err)
	}
}

func newMail() Mail {
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	return Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		FromName:    os.Getenv("MAIL_FROM_NAME"),
		FromAddress: os.Getenv("MAIL_FROM_ADDRESS"),
	}
}
