package network

import (
	"encoding/json"
	"github.com/algo7/tf2_rcon_misc/utils"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type CallbackFunc func(*websocket.Conn)

const wsPath = "/websocket"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigin := "http://localhost:1212"
		return r.Header.Get("Origin") == allowedOrigin
	},
}
var onConnectCallback CallbackFunc

func StartWebsocket(port int, callback CallbackFunc) {
	http.HandleFunc(wsPath, handleWebSocket)
	onConnectCallback = callback

	log.Printf("Starting websocket for IPC communication on path '%s' and port '%d'", wsPath, port)

	// Start your HTTP server
	err := http.ListenAndServe("127.0.0.1:"+strconv.Itoa(port), nil)
	if err != nil {
		log.Panicf("ERROR while creating HTTP server: %v", err)
		return
	}
}

// Send players over the wire as JSON.
func SendPlayers(c *websocket.Conn, players []*utils.PlayerInfo) {
	if len(players) == 0 {
		log.Printf("SendPlayers() Player slice is empty, not sending")
		return
	}

	// Convert the player data to a JSON string
	jsonData, err := json.Marshal(players)
	if err != nil {
		log.Panicf("ERROR while marshalling players as JSON: %v", err)
		return
	}

	//log.Printf("Sending players, json-payload is: %s", string(jsonData))

	if err := c.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		log.Printf("ERROR while sending players as websocket-message: %v", err)
		return
	}
}

// WebSocket handler
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panicf("ERROR while upgrading websocket connection: %v", err)

		// Handle error
		return
	}

	log.Printf("NEW websocket connection from '%s' (requesting: '%s')!", r.RemoteAddr, r.RequestURI)

	// defer Close, ignore error
	defer func(conn *websocket.Conn) {
		_ = conn.Close()
	}(conn)

	onConnectCallback(conn)

	// Handle WebSocket communication here
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			errStr := err.Error()

			// Ignore connection closes
			if strings.Contains(errStr, "close 1005") {
				log.Printf("CLOSED connection from '%s' was closed", r.RemoteAddr)
			} else {
				log.Panicf("ERROR while reading websocket message: %v", err)
			}

			return
		}

		// Process the received message
		processMessage(messageType, p)

		// Example: Echo back the message
		if err := conn.WriteMessage(messageType, p); err != nil {
			return
		}
	}
}

// Process incoming websocket message
func processMessage(messageType int, p []byte) {
	switch messageType {
	case websocket.TextMessage:
		// Handle text message
		text := string(p)
		log.Printf("Received message over websockets: %s", text)
		// Process 'text' as needed

	case websocket.BinaryMessage:
		// Handle binary message
		log.Printf("Received BinaryMessage over websockets.")
		// Process 'p' as needed

	case websocket.CloseMessage:
		// Handle close message
		log.Printf("Received CloseMessage over websockets.")
		// Initiate the WebSocket closing process

	case websocket.PingMessage:
		// Handle ping message
		log.Printf("Received PingMessage over websockets.")
		// Respond with a pong message

	case websocket.PongMessage:
		// Handle pong message
		log.Printf("Received PongMessage over websockets.")
		// Confirm the connection's health

	default:
		log.Printf("Received message over websockets, type: %v - message: %v", messageType, p)
		// Handle other message types or ignore them
	}
}
