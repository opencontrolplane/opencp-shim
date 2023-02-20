package config

// THnis pacake is use to loadf the config file in yaml format and convert it to a []metav1.APIResource

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ApiResource is the struct that holds the api resource config
type ApiResource struct {
	Verbs        []string `yaml:"Verbs"`
	Namespaced   bool     `yaml:"Namespaced"`
	ShortNames   []string `yaml:"ShortNames"`
	Kind         string   `yaml:"Kind"`
	SingularName string   `yaml:"SingularName"`
	Name         string   `yaml:"Name"`
	Version      string   `yaml:"Version"`
}

// GrpcServer is the struct that holds the grpc server config
type GrpcServer struct {
	Host string `yaml:"Host"`
}

// EtcdServer is the struct that holds the etcd server config
type EtcdServer struct {
	Host []string `yaml:"Host"`
}

// Config is the struct that holds the config file
type Config struct {
	ApiResource []ApiResource `yaml:"ApiResource"`
	GrpcServer  GrpcServer    `yaml:"GrpcServer"`
	EtcdServer  EtcdServer    `yaml:"EtcdServer"`
}

// LoadConfig loads the config file and returns a Config struct
func LoadConfig(configFile string) (Config, error) {
	if configFile == "" {
		return Config{}, fmt.Errorf("config file name is empty")
	}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return Config{}, fmt.Errorf("config file %s does not exist", configFile)
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("config file %s does not exist", configFile)
	}

	config := Config{}
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return Config{}, fmt.Errorf("error parsing config file %s", configFile)
	}

	grpcServer := os.Getenv("GRPC_SERVER")
	if grpcServer != "" {
		config.GrpcServer.Host = grpcServer
	}

	return config, nil
}
