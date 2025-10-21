package config

import (
	"github.com/Artexxx/HR-Kafka-QA/library/pg"
	"github.com/Artexxx/HR-Kafka-QA/library/yamlenv"
)

type Config struct {
	Postgres pg.PostgresConfig `yaml:"postgres"`
	Kafka    KafkaConfig       `yaml:"kafka"`
	UserAPI  ApiConfig         `yaml:"userAPI"`
}

type KafkaConfig struct {
	Bootstrap        *yamlenv.Env[string] `yaml:"bootstrap"`
	ProducerClientID *yamlenv.Env[string] `yaml:"producer_client_id"`
	Topics           struct {
		Personal  *yamlenv.Env[string] `yaml:"personal"`
		Positions *yamlenv.Env[string] `yaml:"positions"`
		History   *yamlenv.Env[string] `yaml:"history"`
	} `yaml:"topics"`
}

type ApiConfig struct {
	Port *yamlenv.Env[int] `yaml:"port"`
}
