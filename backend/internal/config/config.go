package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env                  string        `yaml:"env" env:"ENV" env-default:"local"`
	InactiveTimeForAgent int32         `yaml:"inactive_time_for_agent" env-default:"200"`
	TimeForPing          int32         `yaml:"time_for_ping" end-default:"100"`
	TokenTTL             time.Duration `yaml:"tokenTTL" env-default:"1h"`
	JWTSecret            string        `env:"JWT_SECRET" env-required:"true"`
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
	StorageURL        string `yaml:"storage_url" env-default:"postgres://postgres:postgres@localhost:5432/DAEC?sslmode=disable"`
	GooseMigrationDir string `yaml:"goose_migration_dir" env:"GOOSE_MIGRATION_DIR" env-required:"true"`
}

type GRPCServer struct {
	Address                    string `yaml:"address" env-default:"localhost:44044"`
	GRPCClientConnectionString string `yaml:"grpc_client_connection_string" env-default:"auth:44044"`
}

func MustLoad() *Config {
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
