package main

import (
	"fmt"
	"os"
	"time"

	"github.com/evilmagics/dataset_collector/internal/config"
	"github.com/evilmagics/dataset_collector/internal/services"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Generate log filename according to timestamp

	logFilename := fmt.Sprintf("logs_collector_%s.log", time.Now().Format("2006-01-02_15-04-05"))
	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed create log file")
	}

	log.Logger = zerolog.New(zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stderr}, logFile)).With().Timestamp().Logger()
	configPath := config.ParseArgs()

	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed load config")
	}
	log.Info().Str("Path", configPath).Msg("Load config file")

	collector, err := services.NewCollector(conf)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed collect database!")
	}
	collector.CollectAll()

	// time.Sleep(5 * time.Second)
}
