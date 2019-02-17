package main

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type ASyncCallbackServer struct {
	addr   string
	txChan chan string
}

type callbackResponse struct {
	Id     string `json:"id"`
	TxHash string `json:"txHash"`
	Error  string `json:"error"`
}

func NewASyncCallbackServer(addr string, txChan chan string) *ASyncCallbackServer {
	return &ASyncCallbackServer{addr: addr, txChan: txChan}
}

func (s *ASyncCallbackServer) Run() error {
	http.HandleFunc("/sendTransactionAsync", func(w http.ResponseWriter, r *http.Request) {
		d := json.NewDecoder(r.Body)
		p := &callbackResponse{}
		err := d.Decode(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if p.Error != "" {
			http.Error(w, p.Error, http.StatusInternalServerError)
			return
		}
		//		log.WithFields(log.Fields{
		//			"txHash": p.TxHash,
		//		}).Info("POST callback")
		s.txChan <- p.TxHash
	})
	log.WithFields(log.Fields{
		"addr": s.addr,
	}).Info("Starting ASync callback server")
	return http.ListenAndServe(s.addr, nil)
}
