package common

import (
	"github.com/jhopan/Panway/common/env"
)

var UsingSQLite = false
var UsingPostgreSQL = false
var UsingMySQL = false

var SQLitePath = "panway.db"
var SQLiteBusyTimeout = env.Int("SQLITE_BUSY_TIMEOUT", 3000)
