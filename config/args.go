package config

import (
	"flag"
)

func getConfigFilenameFromConsoleArgs() string {
	var filename string
	flag.StringVar(&filename, "config", "", "path to yaml config file")
	flag.Parse()
	return filename
}
