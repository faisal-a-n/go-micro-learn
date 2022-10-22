package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	//Populate requestPayload
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	//Check and validate user
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		fmt.Println("Non existing user", err)
		app.errorJSON(w, errors.New("Invalid credentails"), http.StatusUnauthorized)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if !valid || err != nil {
		fmt.Println("Password mismatch", err)
		app.errorJSON(w, errors.New("Invalid credentails"), http.StatusUnauthorized)
		return
	}

	//Log authentication event
	go app.log(LogPayload{
		Name: "Authentication Event",
		Data: fmt.Sprint("Logged in ", user.Email),
	})

	//Send response
	payload := jsonResponse{
		Code:    200,
		Data:    user,
		Message: "User logged in",
	}
	app.writeJSON(w, http.StatusAccepted, payload, nil)
}

func (app *Config) log(payload LogPayload) error {
	data, _ := json.MarshalIndent(payload, "", "\t")

	logServiceURL := "http://logger-service/log"
	req, _ := http.NewRequest(http.MethodPost, logServiceURL, bytes.NewBuffer(data))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error logging", err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		log.Println("Error logging", err)
		return err
	}
	return nil
}
