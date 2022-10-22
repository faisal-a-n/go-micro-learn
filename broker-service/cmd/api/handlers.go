package main

import (
	"broker-service/event"
	"broker-service/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string      `json:"action" binding:"required,action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"min=6"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From        string   `json:"from"`
	To          string   `json:"to"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	Attachments []string `json:"attachments"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Code:    http.StatusOK,
		Message: "Connected to broker",
	}
	_ = app.writeJSON(w, http.StatusOK, payload, nil)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	switch requestPayload.Action {
	case "authenticate":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.logViaRPC(w, requestPayload.Log)
	case "mail":
		app.mail(w, requestPayload.Mail)
	default:
		app.errorJSON(w, errors.New("action not supported"), http.StatusBadRequest)
	}
}

func makeServiceRequest(payload any, url string) (jsonResponse, error) {
	data, _ := json.MarshalIndent(payload, "", "\t")

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return jsonResponse{}, err
	}
	defer res.Body.Close()

	var serviceResponse jsonResponse
	err = json.NewDecoder(res.Body).Decode(&serviceResponse)
	if err != nil {
		return jsonResponse{}, err
	}
	return serviceResponse, nil
}

func (app *Config) mail(w http.ResponseWriter, payload MailPayload) {
	mailServiceURL := "http://mail-service/send"

	serviceResponse, err := makeServiceRequest(payload, mailServiceURL)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	if serviceResponse.Code != http.StatusOK {
		app.errorJSON(w, errors.New(serviceResponse.Message), serviceResponse.Code)
		return
	}

	app.writeJSON(w, http.StatusOK, serviceResponse)
}

func (app *Config) log(w http.ResponseWriter, payload LogPayload) {
	logServiceURL := "http://logger-service/log"

	serviceResponse, err := makeServiceRequest(payload, logServiceURL)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	if serviceResponse.Code != http.StatusAccepted {
		app.errorJSON(w, errors.New(serviceResponse.Message), serviceResponse.Code)
		return
	}

	app.writeJSON(w, http.StatusOK, serviceResponse)
}

func (app *Config) authenticate(w http.ResponseWriter, payload AuthPayload) {
	//Create json body and send to auth service
	data, _ := json.MarshalIndent(payload, "", "\t")

	//call the service
	req, err := http.NewRequest(http.MethodPost, "http://authentication-service/authenticate", bytes.NewBuffer(data))
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	defer res.Body.Close()

	//validate response
	if res.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("Invalid credentials"), http.StatusUnauthorized)
		return
	}
	if res.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling service"), http.StatusUnauthorized)
		return
	}

	//return response back to client
	var serviceResponse jsonResponse
	err = json.NewDecoder(res.Body).Decode(&serviceResponse)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	if serviceResponse.Code != 200 {
		app.errorJSON(w, errors.New(serviceResponse.Message), http.StatusBadRequest)
		return
	}

	app.writeJSON(w, http.StatusOK, serviceResponse)
}

func (app *Config) logViaRabbitmq(w http.ResponseWriter, payload LogPayload) {
	err := app.pushToQueue(payload)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	var jsonResponse jsonResponse
	jsonResponse.Code = 200
	jsonResponse.Message = "Logged via rabbitmq"
	app.writeJSON(w, http.StatusOK, jsonResponse)
}

func (app *Config) pushToQueue(payload any) error {
	data, _ := json.MarshalIndent(payload, "", "\t")
	message := string(data)

	emitter, err := event.NewEventEmitter(app.rabbitConn)
	if err != nil {
		log.Println("Cannot create emitter", err)
		return err
	}
	err = emitter.Emit(message, "log.INFO")
	if err != nil {
		log.Println("Cannot emit event", err)
		return err
	}
	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logViaRPC(w http.ResponseWriter, payload LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	rpcPayload := RPCPayload{
		Name: payload.Name,
		Data: payload.Data,
	}
	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	app.writeJSON(w, http.StatusOK, jsonResponse{
		Code:    200,
		Message: result,
	})
}

func (app *Config) logViaGRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	conn, err := grpc.Dial("logger-service:50001",
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	defer conn.Close()

	client := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	app.writeJSON(w, http.StatusOK, jsonResponse{
		Code:    200,
		Message: res.Result,
	})
}
