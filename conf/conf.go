// read and expose config
package conf

import (
	"io"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type Config struct {
    DB struct {
        Host string `yaml:"host"`
        Port string `yaml:"port"`
        User string `yaml:"user"`
        Pass string `yaml:"pass"`
        DBName string `yaml:"db_name"`
    } `yaml:"db"`
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
