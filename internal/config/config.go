package config

type Config struct {
	Database struct {
		Driver string `env:"OGSC_DB_DRIVER" env-default:"sqlite3"   yaml:"driver"`
		Dsn    string `env:"OGSC_DB_DSN"    env-default:"./ogsc.db" yaml:"dsn"`
	} `yaml:"database"`

	Server struct {
		Host     string `env:"OGSC_SERVER_HOST" env-default:"127.0.0.1" yaml:"host"`
		Port     int    `env:"OGSC_SERVER_PORT" env-default:"4332"      yaml:"port"`
		LogLevel string `env:"OGSC_LOG_LEVEL"   env-default:"error"     yaml:"log_level"`
		ApiKey   string `env:"OGSC_API_KEY"     env-default:""          yaml:"api_key"`
	} `yaml:"server"`
}
