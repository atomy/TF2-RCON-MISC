package main

import (
	"github.com/gorilla/websocket"
	"log"
	"strings"
	"time"

	"github.com/algo7/tf2_rcon_misc/db"
	"github.com/algo7/tf2_rcon_misc/network"
	"github.com/algo7/tf2_rcon_misc/utils"
)

// Const console message that informs you about forceful autobalance.
//const teamSwitchMessage = "You have switched to team BLU and will receive 500 experience points at the end of the round for changing teams."

// Slice of player info cache struct that holds the player info
var playersInGame []*utils.PlayerInfo

var websocketConnection *websocket.Conn

func main() {
	// Start websocket for IPC with UI-Client
	go network.StartWebsocket(27689, onWebsocketConnectCallback)

	// Init the grok patterns
	utils.GrokInit()

	// Connect to the rcon server
	network.Connect()

	if network.RCONConnection == nil {
		log.Println("Connection to RCON failed")
	}

	// Get the current player name
	res := network.RconExecute("name")
	playerName, err := utils.GrokParsePlayerName(res)

	if err != nil {
		log.Fatalf("%v Please restart the program", err)
	}

	log.Printf("Player Name: %s", playerName)

	// Get log path
	tf2LogPath := utils.LogPathDection()

	// Empty the log file
	err = utils.EmptyLog(tf2LogPath)

	if err != nil {
		log.Fatalf("Unable to empty the log file: %v", err)
	}

	// Tail the log
	log.Println("Tailing Logfile at:", tf2LogPath)
	t, err := utils.TailLog(tf2LogPath)
	if err != nil {
		log.Fatalf("Unable to tail the log file: %v", err)
	}

	// Loop through the text of each received line
	for line := range t.Lines {

		// Parse the line for player info
		playerInfo, err := utils.GrokParse(line.Text)
		if err != nil {
			// log.Printf("GrokParse error: %s at %v", line.Text, err)
		}

		// Parse the line for chat info
		chat, err := utils.GrokParseChat(line.Text)
		if err != nil {
			// log.Printf("GrokParseChat error: %s at %v", line.Text, err)
		}

		// Refresh player list logic
		// Dont assume status headlines as player connects
		if strings.Contains(line.Text, "Lobby updated") || (strings.Contains(line.Text, "connected") && !strings.Contains(line.Text, "uniqueid")) {
			log.Printf("Executing *status* command after line: %s", line.Text)

			// Run the status command when the lobby is updated or a player connects
			network.RconExecute("status")
		}

		if chat != nil {
			log.Printf("Chat: %+v\n", *chat)
		}

		// Save to DB logic
		if playerInfo != nil {

			log.Printf("%+v\n", *playerInfo)

			// Append the player to the player list
			updatePlayers(playerInfo)

			// Create a player document for inserting into MongoDB
			player := db.Player{
				SteamID:   playerInfo.SteamID,
				Name:      playerInfo.Name,
				UpdatedAt: time.Now().UnixNano(),
			}

			// Add the player to the DB
			db.AddPlayer(player)

			// When websocket connected, send over the new players
			if websocketConnection != nil {
				network.SendPlayers(websocketConnection, playersInGame)
			}
		}
	}
}

// Update player collection with supplied new playerInfo entity.
func updatePlayers(playerInfo *utils.PlayerInfo) {
	// Check if the player already exists in the list
	for i, existingPlayer := range playersInGame {
		if existingPlayer.SteamID == playerInfo.SteamID {
			// Player already exists, update the fields
			playersInGame[i] = playerInfo
			return
		}
	}

	playersInGame = append(playersInGame, playerInfo)
}

// Callback that is called once websocket-connection has been established.
func onWebsocketConnectCallback(c *websocket.Conn) {
	websocketConnection = c
	network.SendPlayers(c, playersInGame)
}
