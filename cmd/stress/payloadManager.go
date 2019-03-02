package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type StoreRawRequest struct {
	Payload string `json:"payload"`
	From    string `json:"from"`
}

type StoreRawResponse struct {
	Key string `json:"key"`
}

type TransactionManager struct {
	URL string
	// Stats
}

func (tm *TransactionManager) Send(from, payload string) (string, error) {
	buf := new(bytes.Buffer)
	//TODO below convert RLPtoBase64
	json.NewEncoder(buf).Encode(&StoreRawRequest{
		Payload: payload,
		From:    from,
	})
	url := fmt.Sprintf("%s/storeraw", tm.URL)
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return "", fmt.Errorf("http.NewRequest: %v", err)
	}
	//	req.Header.Set("c11n-to", strings.Join(b64To, ","))
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("client.Do: %v", err)
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var srr StoreRawResponse
	err = dec.Decode(&srr)
	//base64.NewDecoder(base64.StdEncoding, resp.Body)
	// TODO Base64toHex
	if err != nil {
		return "", fmt.Errorf("dec.Decode: %v", err)
	}

	return srr.Key, nil
}

type TMService struct {
	tm       []*TransactionManager
	From     string
	PayloadC chan interface{}
}

func (tms *TMService) DistributePayload(payload string) (string, error) {
	var key string
	for _, tm := range tms.tm {
		_key, err := tm.Send(tms.From, payload)
		if err != nil {
			return "", err
		}
		if key == "" {
			key = _key
		} else if key != _key {
			return "", fmt.Errorf("Key from different between TransactionManager, got %q, want %q", _key, key)
		}
	}
	return key, nil
}

func NewTMService(from string, tmUrls []string) (*TMService, error) {
	var tm []*TransactionManager
	for _, e := range tmUrls {
		_tm := &TransactionManager{URL: e}
		tm = append(tm, _tm)
	}
	return &TMService{
		tm:       tm,
		From:     from,
		PayloadC: make(chan interface{}),
	}, nil
}
