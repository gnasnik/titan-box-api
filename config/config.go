package config

var Cfg Config

type Config struct {
	Mode        string
	ApiListen   string
	DatabaseURL string
	SecretKey   string
}
