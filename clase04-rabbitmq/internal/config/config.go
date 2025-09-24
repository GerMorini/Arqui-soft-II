package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port      string
	Mongo     MongoConfig
	Memcached MemcachedConfig
	RabbitMQ  RabbitMQConfig
}

type MongoConfig struct {
	URI string
	DB  string
}

type MemcachedConfig struct {
	Host       string
	Port       string
	TTLSeconds int
}

type RabbitMQConfig struct {
	Username  string
	Password  string
	QueueName string
	Host      string
	Port      string
}

func Load() Config {
	memcachedTTL, err := strconv.Atoi(getEnv("MEMCACHED_TTL_SECONDS", "60"))
	if err != nil {
		memcachedTTL = 60
	}

	cfg := Config{
		Port: getEnv("PORT", "8080"),
		Mongo: MongoConfig{
			URI: getEnv("MONGO_URI", "mongodb://appuser:apppass@localhost:27017/app?authSource=app"),
			DB:  getEnv("MONGO_DB", "app"),
		},
		Memcached: MemcachedConfig{
			Host:       getEnv("MEMCACHED_HOST", "localhost"),
			Port:       getEnv("MEMCACHED_PORT", "11211"),
			TTLSeconds: memcachedTTL,
		},
		RabbitMQ: RabbitMQConfig{
			Username:  getEnv("RABBITMQ_USER", "admin"),
			Password:  getEnv("RABBITMQ_PASS", "admin"),
			QueueName: getEnv("RABBITMQ_QUEUE_NAME", "items-news"),
			Host:      getEnv("RABBITMQ_HOST", "localhost"),
			Port:      getEnv("RABBITMQ_PORT", "5672"),
		},
	}

	log.Println("========== CONFIGURACIÓN ==========")
	log.Println("PORT:", cfg.Port)
	log.Println("MONGO_URI:", cfg.Mongo.URI)
	log.Println("MONGO_DB:", cfg.Mongo.DB)
	log.Println("MEMCACHED_HOST:", cfg.Memcached.Host)
	log.Println("MEMCACHED_PORT:", cfg.Memcached.Port)
	log.Println("MEMCACHED_TTL_SECONDS:", cfg.Memcached.TTLSeconds)
	log.Println("RABBITMQ_USER:", cfg.RabbitMQ.Username)
	log.Println("RABBITMQ_PASS:", cfg.RabbitMQ.Password)
	log.Println("RABBITMQ_QUEUE_NAME:", cfg.RabbitMQ.QueueName)
	log.Println("RABBITMQ_HOST:", cfg.RabbitMQ.Host)
	log.Println("RABBITMQ_PORT:", cfg.RabbitMQ.Port)
	log.Println("===================================")

	return cfg
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
