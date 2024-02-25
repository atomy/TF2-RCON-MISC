package network

import (
	"github.com/algo7/tf2_rcon_misc/logger"
	"github.com/gorcon/rcon"
	"github.com/gorilla/websocket"
	"net/http"
)

// Create a new instance of the logger.
var log = logger.Logger

// Global variables
var (
	rconHost       string
	RCONConnection *rcon.Conn
)

const (
	rconPort = 27015
)

type CallbackFunc func(*websocket.Conn)

type Message struct {
	Type string `json:"type"`
}

const wsPath = "/websocket"

var HttpServer *http.Server // Exported by capitalizing the first letter

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var onConnectCallback CallbackFunc
