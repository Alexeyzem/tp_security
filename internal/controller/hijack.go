package controller

import (
	"net"
	"net/http"
)

func hijackClientConnection(w http.ResponseWriter) (net.Conn, error) {}
