package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/algo7/tf2_rcon_misc/commands"
	"github.com/algo7/tf2_rcon_misc/db"
	"github.com/algo7/tf2_rcon_misc/network"
	"github.com/algo7/tf2_rcon_misc/utils"
)

// Const console message that informs you about forceful auto-balance.
//const teamSwitchMessage = "You have switched to team BLU and will receive 500 experience points at the end of the round for changing teams."

// playersInGame is a slice of player info cache struct that holds the player info
var playersInGame []*utils.PlayerInfo

// Holds the last tf_lobby_debug response for usage.
var lastLobbyDebugResponse string
var lastUpdate int64
var currentPlayer string

var websocketConnection *websocket.Conn

func main() {
	signals := setupSignalHandler()

	// Goroutine to handle signals
	go func() {
		sig := <-signals
		fmt.Println("Signal received:", sig)

		// Notify the main logic to shut down
		// Perform cleanup or shutdown procedures
		fmt.Println("Performing graceful shutdown.")

		if network.HttpServer != nil {
			_ = network.HttpServer.Close()
			fmt.Println("HttpServer closed.")
		}

		os.Exit(0)
	}()

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

	err := error(nil)
	currentPlayer, err = utils.GrokParsePlayerName(res)

	if err != nil {
		log.Fatalf("%v Please restart the program", err)
	}

	log.Printf("Current player is %s", currentPlayer)

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

	// Start player watcher.
	go startUpdatePlayerWatcher()

	// Loop through the text of each received line
	for line := range t.Lines {

		// Refresh player list logic
		// Don't assume status headlines as player connects
		if strings.Contains(line.Text, "Lobby updated") || (strings.Contains(line.Text, "connected") && !strings.Contains(line.Text, "uniqueid")) {
			log.Printf("Executing *status* command after line: %s", line.Text)

			// Clear the player list
			playersInGame = []*utils.PlayerInfo{}
			// Run the status command when the lobby is updated or a player connects
			network.RconExecute("status")
		}

		// Parse the line for player info
		if playerInfo, err := utils.GrokParse(line.Text); err == nil {
			// log.Printf("%+v\n", *playerInfo)

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

		// Parse the line for chat info
		if chat, err := utils.GrokParseChat(line.Text); err == nil {

			log.Printf("Chat: %+v\n", *chat)

			// Parse the chat message for commands
			if command, args, err := utils.GrokParseCommand(chat.Message); err == nil {
				commands.CommandExecuted(command, args, chat.PlayerName, currentPlayer)
			}

			// Get the player's steamID64 from the playersInGame
			steamID, err := utils.GetSteamIDFromPlayerName(chat.PlayerName, playersInGame)

			if err == nil {
				// Create a chat document for inserting into MongoDB
				chatInfo := db.Chat{
					SteamID:   steamID,
					Name:      chat.PlayerName,
					Message:   chat.Message,
					UpdatedAt: time.Now().UnixNano(),
				}
				db.AddChat(chatInfo)
			}
		}
	}
}

// Update player collection with supplied new playerInfo entity.
func updatePlayers(playerInfo *utils.PlayerInfo) {
	var lobbyPlayers []utils.LobbyDebugPlayer

	if "Failed to find lobby shared object" != lastLobbyDebugResponse {
		lobbyPlayers = utils.ParseLobbyResponse(lastLobbyDebugResponse)
	}

	// Find ourselves and set flag to true.
	if playerInfo.Name == currentPlayer {
		playerInfo.IsMe = true
	} else {
		playerInfo.IsMe = false
	}

	lobbyPlayer := utils.FindLobbyPlayerBySteamId(lobbyPlayers, playerInfo.SteamID)

	if lobbyPlayer != nil {
		//log.Printf("dbg: %+v\n", lobbyPlayer)
		playerInfo.Team = lobbyPlayer.Team
		playerInfo.Type = lobbyPlayer.Type
		playerInfo.MemberType = lobbyPlayer.MemberType
	}

	// Check if the player already exists in the list
	for i, existingPlayer := range playersInGame {
		if existingPlayer.SteamID == playerInfo.SteamID {
			// Player already exists, update the fields
			// Preserve tf-lobby-fields if new ones are empty
			if len(playerInfo.Team) <= 0 {
				playerInfo.Team = playersInGame[i].Team
			}

			if len(playerInfo.Type) <= 0 {
				playerInfo.Team = playersInGame[i].Type
			}

			if len(playerInfo.MemberType) <= 0 {
				playerInfo.Team = playersInGame[i].MemberType
			}

			playersInGame[i] = playerInfo
			return
		}
	}

	// If playerInfo is $ME, mark it as me.
	if playerInfo.Name == currentPlayer {
		playerInfo.IsMe = true
	}

	playersInGame = append(playersInGame, playerInfo)
	lastUpdate = time.Now().Unix()
}

// onWebsocketConnectCallback Callback that is called once websocket-connection has been established.
func onWebsocketConnectCallback(c *websocket.Conn) {
	websocketConnection = c
	network.SendPlayers(c, playersInGame)
}

// startUpdatePlayerWatcher Initializes player updates every 10 seconds if there have been none.
func startUpdatePlayerWatcher() {
	for {
		// Sleep for 10 seconds
		time.Sleep(10 * time.Second)

		// Check when last update happened.
		if (lastUpdate + 10) < time.Now().Unix() {
			log.Println("Executing *status* + *tf_lobby_debug* command after scheduled 10s")
			lastLobbyDebugResponse = network.RconExecute("tf_lobby_debug")
			network.RconExecute("status")
		} else {
			log.Printf("No update necessary, last one happened '%d' seconds ago!\n", time.Now().Unix()-lastUpdate)
		}
	}
}

// setupSignalHandler Sets up a signal handler to termination from the outside.
func setupSignalHandler() chan os.Signal {
	// Setting up channel to listen for signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	// Channel to notify the main logic to shut down
	return signals
}
