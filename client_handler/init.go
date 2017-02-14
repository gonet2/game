package client_handler

import (
	"game/db"
	"time"
)

var (
	DefaultDatabase db.Database
)

func Init(mongodb string, concurrent int, timeout time.Duration) {
	DefaultDatabase.Init(mongodb, concurrent, timeout)
}
