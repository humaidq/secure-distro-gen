package config

import (
	"bytes"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

var (
	// WorkingDir is the current working directory of the project.
	WorkingDir string
	// ConfigPath is the configuration file name.
	ConfigPath = "config.toml"
	// Config is where the current configuration is loaded.
	Config Configuration
	// StartTime is the time when the server started.
	StartTime = time.Now()
	// Logger is the zap logger.
	Logger *zap.SugaredLogger
)

// Configuration represents the configuration file format.
type Configuration struct {
	SiteName    string // SiteName is the name of the site.
	SitePort    string // SitePort is the port to run the web server on.
	OrigISOFile string // OrigISOFile is the base file used to generate the system.
	ISOFileMD5  string // ISOFileMD5 is the md5sum of the OrigISOFile.
}

func newConfig() Configuration {
	return Configuration{
		SiteName:    "CSLDG",
		SitePort:    "8080",
		OrigISOFile: "base.iso",
		ISOFileMD5:  "7f3c4465618e17104f0c76e5646c7caccb4161bc01a102ed04d34b1b4f1e4208",
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	var err error
	WorkingDir, err = os.Getwd()
	if err != nil {
		log.Fatal("Cannot get working directory!", err)
	}
}

func SetupLogger() {
	logger, _ := zap.NewDevelopment()
	Logger = logger.Sugar()
}

// LoadConfig loads the configuration file from disk. It will also generate one
// if it doesn't exist.
func LoadConfig() {
	var err error
	if _, err = toml.DecodeFile(WorkingDir+"/"+ConfigPath, &Config); err != nil {
		log.Printf("Cannot load config file. Error: %s", err)
		if os.IsNotExist(err) {
			log.Println("Generating new configuration file, as it doesn't exist")
			var err error

			buf := new(bytes.Buffer)
			if err = toml.NewEncoder(buf).Encode(newConfig()); err != nil {
				log.Fatal(err)
			}

			err = ioutil.WriteFile(ConfigPath, buf.Bytes(), 0600)
			if err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}
	}
}

// SaveConfig saves the configuration from memory to disk.
func SaveConfig() error {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(Config); err != nil {
		return err
	}

	if err := ioutil.WriteFile(ConfigPath, buf.Bytes(), 0600); err != nil {
		return err
	}
	return nil
}
