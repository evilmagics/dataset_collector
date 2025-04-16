package main

import (
	"os"

	"github.com/evilmagics/dataset_collector/internal/config"
	"github.com/evilmagics/dataset_collector/internal/services"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
)

func main() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	configPath := config.ParseArgs()

	fs := afero.NewOsFs()

	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed load config")
	}

	datasetConf, err := config.LoadDataset(fs, configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed load config")
	}

	collector, err := services.NewCollector(conf, datasetConf)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed collect database!")
	}
	collector.CollectAll()
}
