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

	Sharding struct {
		UserShardN      int `yaml:"user_shard_n"`
		UserPhoneShardN int `yaml:"user_phone_shard_n"`
		UserEmailShardN int `yaml:"user_email_shard_n"`
	} `yaml:"sharding"`

	App struct {
		Port              string        `yaml:"port"`
		AuthTokenExpire   time.Duration `yaml:"auth_token_expire"`
		CheckUserIsAuthor bool          `yaml:"check_user_is_author"`
		DefaultPageSize   int           `yaml:"default_page_size"`
		MaxPageSize       int           `yaml:"max_page_size"`
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
	return config
}
