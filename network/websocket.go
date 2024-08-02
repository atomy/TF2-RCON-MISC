package network

import (
	"encoding/json"
	"github.com/algo7/tf2_rcon_misc/utils"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// StartWebsocket Startup websocket server for communication with electron UI.
func StartWebsocket(port int, callback CallbackFunc) {
	http.HandleFunc(wsPath, handleWebSocket)
	onConnectCallback = callback

	HttpServer = &http.Server{Addr: "127.0.0.1:" + strconv.Itoa(port)}

	log.Printf("Starting websocket for IPC communication on path '%s' and port '%d'", wsPath, port)

	// Start your HTTP server
	if err := HttpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Panicf("ERROR while creating HTTP server: %v", err)
		return
	}
}

// SendPlayers Send players encapsulated in a player-update over the wire as JSON.
func SendPlayers(c *websocket.Conn, players []*utils.PlayerInfo) {
	if len(players) == 0 {
		// log.Printf("SendPlayers() Player slice is empty, not sending")
		return
	}

	playerUpdate := utils.PlayerUpdate{
		Type:           "player-update",
		CurrentPlayers: players,
	}

	// Convert the player data to a JSON string
	jsonData, err := json.Marshal(playerUpdate)
	if err != nil {
		log.Panicf("ERROR while marshalling players as JSON: %v", err)
		return
	}

	// log.Printf("Sending players, json-payload is: %s", string(jsonData))

	if err := c.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		log.Printf("ERROR while sending players as websocket-message: %v", err)
		return
	}
}

// SendFrag, send new frag entries over the network
func SendFrag(c *websocket.Conn, frag *utils.FragInfo) {
	fragWsInfo := utils.FragWsInfo{
		Type: "frag",
		Frag: frag,
	}

	// Convert the frag data to a JSON string
	jsonData, err := json.Marshal(fragWsInfo)
	if err != nil {
		log.Panicf("ERROR while marshalling frag as JSON: %v", err)
		return
	}

	//log.Printf("Sending frag, json-payload is: %s", string(jsonData))

	if err := c.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		log.Printf("ERROR while sending frag as websocket-message: %v", err)
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
			if strings.Contains(errStr, "close ") {
				log.Printf("CLOSED connection from '%s' was closed", r.RemoteAddr)
			} else {
				log.Panicf("ERROR while reading websocket message: %v", err)
			}

			return
		}

		// Process the received message
		processRawMessage(messageType, p)

		// Example: Echo back the message
		if err := conn.WriteMessage(messageType, p); err != nil {
			return
		}
	}
}

// Process incoming websocket message
func processRawMessage(messageType int, p []byte) {
	switch messageType {
	case websocket.TextMessage:
		// Handle text message
		text := string(p)
		log.Printf("Received message over websockets: %s", text)

		// Decode JSON
		var msg Message
		err := json.Unmarshal([]byte(text), &msg)

		if err != nil {
			log.Printf("Error decoding JSON: %s", err)
			return
		}

		processJsonMessage(msg)
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

// processJsonMessage Process incomming message over websockets that has already been json-decoded into a struct.
func processJsonMessage(msg Message) {
	// Exit message, telling us to shut down.
	if msg.Type == "exit" {
		os.Exit(0)
	}
}
