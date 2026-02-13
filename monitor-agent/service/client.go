package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func postJSON(url string, payload interface{}) (*http.Response, error) {
	data, _ := json.Marshal(payload)
	return httpClient.Post(url, "application/json", bytes.NewBuffer(data))
}
