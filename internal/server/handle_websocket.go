package server

import (
	"net/http"

	"nhooyr.io/websocket"
)

func (s *Server) WebsocketHandler(w http.ResponseWriter, r *http.Request) {

	// perform any checks to allow a connection or deny


	// all ok, start a websocket
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // HTTP y'all...
		// Subprotocols:       []string{}, // if you want to negotiate different protocol behavior
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer c.Close(websocket.StatusAbnormalClosure, "ruh roh")
	s.Log.WithField("subprotocol", c.Subprotocol()).Info("Adding live socket listener")


	// If you don't want to read anything from the client:
	/*
	// let the system read pings, keepalives, etc. we have no info to read here
	done := c.CloseRead(r.Context())
	*/

	
	// do some writes?

	// .. ok g'bye?
	<-r.Context().Done()
}

