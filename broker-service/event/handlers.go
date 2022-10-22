package event

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func logEvent(payload Payload) error {
	logServiceURL := "http://logger-service/log"

	err := makeServiceRequest(payload, logServiceURL)
	if err != nil {
		return err
	}
	return nil
}

func makeServiceRequest(payload any, url string) error {
	data, _ := json.MarshalIndent(payload, "", "\t")

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
