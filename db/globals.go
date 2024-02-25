package db

import (
	"github.com/algo7/tf2_rcon_misc/logger"
	"os"
)

// Create a new instance of the logger.
var log = logger.Logger

// Connect to the DB
var client = connect()

var (
	// Database  name
	mongoDBName = os.Getenv("MONGODB_NAME")
)

// Player document struct
type Player struct {
	SteamID   int64  `bson:"SteamID"`
	Name      string `bson:"Name"`
	UpdatedAt int64  `bson:"UpdatedAt"`
}

// Chat document struct
type Chat struct {
	SteamID   int64  `bson:"SteamID"`
	Name      string `bson:"Name"`
	Message   string `bson:"message,omitempty"`
	UpdatedAt int64  `bson:"updatedAt"`
}
