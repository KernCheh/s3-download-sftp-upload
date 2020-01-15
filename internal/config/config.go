package config

import (
	"sync"

	"github.com/kelseyhightower/envconfig"
)

// Config is a structure which contains all environment variables for the application
// Usage: `config.GetInstance().<variable>``, e.g `config.GetInstance().Stage`
type Config struct {
	Stage string `envconfig:"STAGE" default:"dev"`

	AuthorizedConfigID string `envconfig:"AUTHORIZED_CONFIG_ID"`

	SftpHost     string `envconfig:"SFTP_HOST"`
	SftpPort     string `envconfig:"SFTP_PORT" default:"22"`
	SftpUserName string `envconfig:"SFTP_USERNAME"`
	SftpPassword string `envconfig:"SFTP_PASSWORD"`

	UploadPath string `envconfig:"UPLOAD_PATH" default:"/import-inbox/"`
}

var instance *Config
var once sync.Once

// GetInstance returns a Config pointer to retrieve environment variables
func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{}

		if err := envconfig.Process("", instance); err != nil {
			panic(err)
		}
	})
	return instance
}
