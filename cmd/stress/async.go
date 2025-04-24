// Copyright (C) 2019 Sylvain 6120 Laurent
// This file is part of eth-stress <https://github.com/Magicking/eth-stress>.
//
// eth-stress is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// eth-stress is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with eth-stress.  If not, see <http://www.gnu.org/licenses/>.

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
		log.WithFields(log.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		}).Debug("Received callback request")

		d := json.NewDecoder(r.Body)
		p := &callbackResponse{}
		err := d.Decode(p)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to decode callback payload")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if p.Error != "" {
			log.WithFields(log.Fields{
				"error": p.Error,
				"id":    p.Id,
			}).Error("Callback reported error")
			http.Error(w, p.Error, http.StatusInternalServerError)
			return
		}

		log.WithFields(log.Fields{
			"id":     p.Id,
			"txHash": p.TxHash,
		}).Info("Successfully processed callback")

		s.txChan <- p.TxHash
	})
	log.WithFields(log.Fields{
		"addr": s.addr,
	}).Info("Starting ASync callback server")
	return http.ListenAndServe(s.addr, nil)
}
