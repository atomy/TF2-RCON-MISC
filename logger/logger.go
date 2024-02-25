package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"time"
)

// AppLogger is a custom logger type
type AppLogger struct {
	*log.Logger
}

var (
	// Logger is the global logger instance
	Logger *AppLogger
)

// LogMessage is a struct for log-messages over websockets, it has its dedicated type
type LogMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

var wsConnection *websocket.Conn

// init initialize the logger
func init() {
	// Create a default logger
	defaultLogger := log.New(os.Stdout, "", 0)

	// Initialize the global logger
	Logger = &AppLogger{Logger: defaultLogger}
}

// Log logs the given message with a timestamp
func (l *AppLogger) Log(m string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := "[" + timestamp + "] " + m
	l.Printf("%s", msg)
	sendLogs(wsConnection, msg)
}

// Printf formats and logs the given message with a timestamp
func (l *AppLogger) Printf(format string, v ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	var buf bytes.Buffer
	buf.WriteString("[" + timestamp + "] ")
	buf.WriteString(format)
	l.Logger.Printf(buf.String(), v...)
	sendLogs(wsConnection, buf.String())
}

// Println logs the given message with a timestamp
func (l *AppLogger) Println(v ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	var buf bytes.Buffer
	buf.WriteString("[" + timestamp + "] ")
	buf.WriteString(fmt.Sprintln(v...))
	l.Logger.Print(buf.String())
	sendLogs(wsConnection, buf.String())
}

// SetWsConnection set package-global wsConnection var when it comes available
func (l *AppLogger) SetWsConnection(connection *websocket.Conn) {
	wsConnection = connection
}

// sendLogs send the log message over ws if ws-connection is present
func sendLogs(ws *websocket.Conn, s string) {
	if ws != nil {
		sendLogMessage(ws, s)
	}
}

// sendLogMessage sends given log message over provided ws connection
func sendLogMessage(c *websocket.Conn, s string) {
	if len(s) == 0 {
		// log.Printf("SendPlayers() Player slice is empty, not sending")
		return
	}

	logMessage := LogMessage{
		Type:    "application-log",
		Message: s,
	}

	// Convert the player data to a JSON string
	jsonData, err := json.Marshal(logMessage)
	if err != nil {
		log.Panicf("ERROR while marshalling log-message as JSON: %v", err)
		return
	}

	//fmt.Printf("Sending log-message, json-payload is: %s", string(jsonData))

	if err := c.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		fmt.Printf("ERROR while sending players as websocket-message: %v", err)
		return
	}
}
