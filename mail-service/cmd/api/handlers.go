package main

import (
	"errors"
	"log"
	"net/http"
)

type mailMessageRequest struct {
	From        string   `json:"from"`
	To          string   `json:"to"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	Attachments []string `json:"attachments"`
}

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	var payload mailMessageRequest
	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	err = app.Mailer.SendSMTPMessage(Message{
		From:        payload.From,
		To:          payload.To,
		Subject:     payload.Subject,
		Data:        payload.Body,
		Attachments: payload.Attachments,
	})
	if err != nil {
		log.Println("Error sending mail", err)
		app.errorJSON(w, errors.New("Error in sending mail"), http.StatusInternalServerError)
		return
	}
	var response jsonResponse
	response.Code = 200
	response.Message = "Mail has been sent!"
	app.writeJSON(w, http.StatusOK, response)
}
