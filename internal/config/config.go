package config

import (
	"HR-Kafka-QA/library/pg"
	"HR-Kafka-QA/library/yamlenv"
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

var (
	CorsAllowHeaders = "Access-Control-Allow-Origin, Access-Control-Allow-Methods, Access-Control-Max-Age, Access-Control-Allow-Credentials, Content-Type, Authorization, Origin, X-Requested-With , Accept"
	CorsAllowMethods = "HEAD, GET, POST, PUT, DELETE, OPTIONS"
	CorsAllowOrigin  = "*"
)
