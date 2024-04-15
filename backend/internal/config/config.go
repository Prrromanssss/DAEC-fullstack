package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env              string `yaml:"env" env:"ENV" env-default:"local"`
	DatabaseInstance `yaml:"database_instance" env-required:"true"`
	RabbitQueue      `yaml:"rabbit_queue" env-required:"true"`
	HTTPServer       `yaml:"http_server" env-required:"true"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type RabbitQueue struct {
	RabbitMQURL               string `yaml:"rabbitmq_url" env-default:"amqp://guest:guest@localhost:5672/"`
	QueueForSendToAgents      string `yaml:"queue_for_send_to_agents" env-required:"true"`
	QueueForConsumeFromAgents string `yaml:"queue_for_consume_from_agents" env-required:"true"`
}

type DatabaseInstance struct {
	StorageURL        string `yaml:"storage_url" env-default:"postgres://postgres:postgres@localhost:5432/daee?sslmode=disable"`
	GooseMigrationDir string `yaml:"goose_migration_dir" env:"GOOSE_MIGRATION_DIR" env-required:"true"`
}

func MustLoad() *Config {
	err := godotenv.Load("local.env")
	if err != nil {
		log.Fatalf("Can't parse env file: %v", err)
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
