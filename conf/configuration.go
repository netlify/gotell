package conf

import (
	"bufio"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Configuration struct {
	Threads struct {
		Source      string `mapstructure:"source" json:"source"`
		Destination string `mapstructure:"destination" json:"destination"`
		Host        string `mapstructure:"host" json:"host"`
		Port        int    `mapstructure:"port" json:"port"`
	}

	API struct {
		SiteURL     string `mapstructure:"site_url" json:"site_url"`
		Repository  string `mapstructure:"repository" json:"repository"`
		AccessToken string `mapstructure:"access_token" json:"access_token"`
		Host        string `mapstructure:"host" json:"host"`
		Port        int    `mapstructure:"port" json:"port"`
	} `mapstructure:"api" json:"api"`

	Logging struct {
		Level string `mapstructure:"level" json:"level"`
		File  string `mapstructure:"file" json:"file"`
	} `mapstructure:"logging" json:"logging"`
}

func LoadConfig(cmd *cobra.Command) (*Configuration, error) {
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return nil, err
	}

	viper.SetEnvPrefix("gotell")
	viper.SetDefault("threads.source", "threads")
	viper.SetDefault("threads.destination", "dist")
	viper.SetDefault("threads.port", "9091")

	if os.Getenv("PORT") == "" {
		viper.SetDefault("api.port", "9090")
	}

	viper.SetConfigType("json")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if configFile, _ := cmd.Flags().GetString("config"); configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("./")
		viper.AddConfigPath("$HOME/.netlify/gotell/")
	}

	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	config := new(Configuration)
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	if err := populateConfig(config); err != nil {
		return nil, err
	}

	if err := configureLogging(config); err != nil {
		return nil, err
	}

	return config, nil
}

// configureLogging will take the logging configuration and also adds
// a few default parameters
func configureLogging(config *Configuration) error {
	logConfig := config.Logging

	// use a file if you want
	if logConfig.File != "" {
		f, errOpen := os.OpenFile(logConfig.File, os.O_RDWR|os.O_APPEND, 0660)
		if errOpen != nil {
			return errOpen
		}
		logrus.SetOutput(bufio.NewWriter(f))
	}

	if logConfig.Level != "" {
		level, err := logrus.ParseLevel(strings.ToUpper(logConfig.Level))
		if err != nil {
			return err
		}
		logrus.SetLevel(level)
	}

	// always use the fulltimestamp
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: false,
	})

	return nil
}
