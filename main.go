package main

import (
	"flag"
	"github.com/finishy1995/mongo-adapter/base"
	"github.com/finishy1995/mongo-adapter/library/log"
	"github.com/finishy1995/mongo-adapter/network"
	"github.com/finishy1995/mongo-adapter/processor/query"
	"gopkg.in/yaml.v2"
	"os"
)

var configFile = flag.String("f", "config.yaml", "the config file")

type Config struct {
	Log            string               `yaml:"log"`
	Host           string               `yaml:"host"`
	Authentication AuthenticationConfig `yaml:"auth"`
	Mongo          MongoConfig          `yaml:"mongo"`
}

type MongoConfig struct {
	Host string `yaml:"host"`
}

type AuthenticationConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Salt     string `yaml:"salt"`
}

func main() {
	flag.Parse()

	// Read the YAML configuration file
	data, err := os.ReadFile(*configFile)
	if err != nil {
		panic(err)
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}

	switch config.Log {
	case "debug":
		log.SetLevel(log.DEBUG)
	case "info":
		log.SetLevel(log.INFO)
	case "warn":
		log.SetLevel(log.WARNING)
	case "error":
		log.SetLevel(log.ERROR)
	default:
		log.SetLevel(log.INFO)
	}

	if config.Mongo.Host == "" {
		config.Mongo.Host = "localhost:27017"
	}
	if config.Host == "" {
		config.Host = "localhost:28017"
	}
	if config.Authentication.Username == "" {
		config.Authentication.Username = "admin"
	}
	if config.Authentication.Password == "" {
		config.Authentication.Password = "admin"
	}
	if config.Authentication.Salt == "" {
		config.Authentication.Salt = "mongo-adapter"
	}
	base.MustSetupMongoClient(config.Mongo.Host)
	query.MustRegisterAuthentication(config.Authentication.Username, config.Authentication.Password, config.Authentication.Salt)
	network.NewServerAndMustStart(config.Host)
}
