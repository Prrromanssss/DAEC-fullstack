package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env                  string        `yaml:"env" env:"ENV" env-default:"local"`
	LogPathAgent         string        `yaml:"log_path_agent" env-required:"true"`
	LogPathOrchestrator  string        `yaml:"log_path_orchestrator" env-required:"true"`
	LogPathAuth          string        `yaml:"log_path_auth" env-required:"true"`
	InactiveTimeForAgent int32         `yaml:"inactive_time_for_agent" env-default:"200"`
	TimeForPing          int32         `yaml:"time_for_ping" end-default:"100"`
	TokenTTL             time.Duration `yaml:"tokenTTL" env-default:"1h"`
	GRPCServer           `yaml:"grpc_server" env-required:"true"`
	DatabaseInstance     `yaml:"database_instance" env-required:"true"`
	RabbitQueue          `yaml:"rabbit_queue" env-required:"true"`
	HTTPServer           `yaml:"http_server" env-required:"true"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type RabbitQueue struct {
	RabbitMQURL                 string `yaml:"rabbitmq_url" env-default:"amqp://guest:guest@localhost:5672/"`
	QueueForExpressionsToAgents string `yaml:"queue_for_expressions_to_agents" env-required:"true"`
	QueueForResultsFromAgents   string `yaml:"queue_for_results_from_agents" env-required:"true"`
}

type DatabaseInstance struct {
	StorageURL        string `yaml:"storage_url" env-default:"postgres://postgres:postgres@localhost:5432/daee?sslmode=disable"`
	GooseMigrationDir string `yaml:"goose_migration_dir" env:"GOOSE_MIGRATION_DIR" env-required:"true"`
}

type GRPCServer struct {
	Address string `yaml:"address" env-default:"localhost:44044"`
}

func MustLoad() *Config {
	path, err := os.Getwd()
	if err != nil {
		log.Fatalf("can't get pwd: %v", err)
	}

	err = godotenv.Load(fmt.Sprintf("%s/local.env", filepath.Dir(filepath.Dir(filepath.Dir(path)))))
	if err != nil {
		log.Fatalf("can't parse env file: %v", err)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {

		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
