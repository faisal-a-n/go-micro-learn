package main

import (
	"logger-service/data"
	"net/http"
	"time"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var payload JSONPayload
	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	currentTime := time.Now()
	event := data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}
	err = app.Models.LogEntry.Insert(event)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	res := jsonResponse{
		Code:    200,
		Message: "Event logged",
		Data:    event,
	}
	app.writeJSON(w, http.StatusAccepted, res)
}
