// read and expose config
package conf

import (
	"io"
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v3"
)

var Global *Config

type Config struct {
	DB struct {
		Type string `yaml:"type"`

		MySQL struct {
			Host   string `yaml:"host"`
			Port   string `yaml:"port"`
			User   string `yaml:"user"`
			Pass   string `yaml:"pass"`
			DBName string `yaml:"db_name"`
		} `yaml:"mysql"`

		Sqlite3 struct {
			DSN string `yaml:"dsn"`
		} `yaml:"sqlite3"`

		AutoMigrate bool `yaml:"auto_migrate"`
	} `yaml:"db"`

	Redis struct {
		Addr     string `yaml:"addr"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`

	Sharding struct {
		UserShardN int `yaml:"user_shard_n"`
	} `yaml:"sharding"`

	App struct {
		Port              string `yaml:"port"`
		CheckUserIsAuthor bool   `yaml:"check_user_is_author"`
		DefaultPageSize   int    `yaml:"default_page_size"`
		MaxPageSize       int    `yaml:"max_page_size"`
		Expire            struct {
			AuthToken time.Duration `yaml:"auth_token_expire"`
			UserInfo  time.Duration `yaml:"user_info"`
			PostInfo  time.Duration `yaml:"post_info"`
		} `yaml:"expire"`
		Timeout struct {
			Default time.Duration `yaml:"default"`
		} `yaml:"timeout"`
		BcrytpCost int `yaml:"bcrypt_cost"`
	} `yaml:"app"`
}

func FromYAML(r io.Reader) Config {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatalf("fails to read config from yaml: %v\n", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("fails to unmarshal config from yaml: %v\n", err)
	}
	assigneDefaults(&config)
	return config
}

func assigneDefaults(config *Config) {
	if config.App.Timeout.Default <= 0 {
		// default timeout is 1 min
		config.App.Timeout.Default = time.Minute
		log.Println("Use default timeout: 1 min")
	}
}
