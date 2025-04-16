package config

import "flag"

func ParseArgs() string {
	conf := flag.String("config", "config.yaml", "config file path")

	flag.Parse()
	return *conf
}
