package common

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/common/logger"
)

var (
	Port         = flag.Int("port", 3000, "the listening port")
	PrintVersion = flag.Bool("version", false, "print version and exit")
	PrintHelp    = flag.Bool("help", false, "print help and exit")
	LogDir       = flag.String("log-dir", "./logs", "specify the log directory")
)

func printHelp() {
	fmt.Println("Panway " + Version + " - AI API gateway for selling tokens.")
	fmt.Println("Copyright (C) 2023 JustSong (original One API). Panway modifications (C) 2026 jhopan.")
	fmt.Println("GitHub: https://github.com/jhopan/Panway")
	fmt.Println("Usage: panway [--port <port>] [--log-dir <log directory>] [--version] [--help]")
}

func Init() {
	flag.Parse()

	if *PrintVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	if *PrintHelp {
		printHelp()
		os.Exit(0)
	}

	secret := os.Getenv("SESSION_SECRET")
	switch {
	case secret == "":
		if os.Getenv("GIN_MODE") == gin.ReleaseMode {
			log.Fatal("SESSION_SECRET is empty: set it to a long random string before running in release mode, " +
				"otherwise every restart invalidates all sessions")
		}
		logger.SysError("SESSION_SECRET is empty: using a random secret, so all sessions are invalidated on every restart")
	case secret == "random_string" || secret == "ubah-ke-string-acak-panjang":
		logger.SysError("SESSION_SECRET is set to an example value, please change it to a random string.")
	case len(secret) < 32:
		logger.SysError("SESSION_SECRET is shorter than 32 characters; use a longer random string.")
		config.SessionSecret = secret
	default:
		config.SessionSecret = secret
	}
	if os.Getenv("SQLITE_PATH") != "" {
		SQLitePath = os.Getenv("SQLITE_PATH")
	}
	if *LogDir != "" {
		var err error
		*LogDir, err = filepath.Abs(*LogDir)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := os.Stat(*LogDir); os.IsNotExist(err) {
			err = os.Mkdir(*LogDir, 0777)
			if err != nil {
				log.Fatal(err)
			}
		}
		logger.LogDir = *LogDir
	}
}
