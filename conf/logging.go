package conf

import (
	"bufio"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
)

// LoggingConfig specifies all the parameters needed for logging
type LoggingConfig struct {
	Level string
	File  string
}

// ConfigureLogging will take the logging configuration and also adds
// a few default parameters
func ConfigureLogging(config *LoggingConfig) (*logrus.Entry, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	// use a file if you want
	if config.File != "" {
		f, errOpen := os.OpenFile(config.File, os.O_RDWR|os.O_APPEND, 0660)
		if errOpen != nil {
			return nil, errOpen
		}
		logrus.SetOutput(bufio.NewWriter(f))
	}

	if config.Level != "" {
		level, err := logrus.ParseLevel(strings.ToUpper(config.Level))
		if err != nil {
			return nil, err
		}
		logrus.SetLevel(level)
	}

	// always use the fulltimestamp
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: false,
	})

	return logrus.StandardLogger().WithField("hostname", hostname), nil
}
