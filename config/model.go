package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type ConfigurationEnv string

const (
	DevEnv ConfigurationEnv = "dev"
	PrdEnv ConfigurationEnv = "prd"
)

func convert2Env(env string) (ConfigurationEnv, error) {
	switch env {
	case "dev":
		return DevEnv, nil
	case "prd":
		return PrdEnv, nil
	default:
		return "", errors.New("illegal env, should be dev or prd")
	}
}

type Configuration map[ConfigurationEnv]*EachConfig

type EachConfig struct {
	AppName         string       `yaml:"app_name"`
	Host            string       `yaml:"host"`
	Port            int          `yaml:"port"`
	CmdsScriptsPath string       `yaml:"cmds_scripts_path"`
	MySqlConfig     *MySqlConfig `yaml:"mysql_config"`
	RedisConfig     *RedisConfig `yaml:"redis_config"`

	Env ConfigurationEnv
}

type MySqlConfig struct {
	Addr     string `yaml:"addr"`
	UserName string `yaml:"user_name"`
	Pwd      string `yaml:"pwd"`
	DBName   string `yaml:"db_name"`
}

type RedisConfig struct {
	Addr string `yaml:"addr"`
}

type args struct {
	ConfigPath string
	Env        ConfigurationEnv
	Port       int
}

func parse(args *args) *EachConfig {
	if args.ConfigPath == "" {
		panic("CONFIG_PATH not set, plz check your environs or cmd args")
	}
	bs, err := ioutil.ReadFile(args.ConfigPath)
	if err != nil {
		log.Fatalf("ConfigForEnv parse failed, read file failed, err=[%v]", err)
	}
	configs := make(Configuration)
	err = yaml.Unmarshal(bs, &configs)
	if err != nil {
		log.Fatalf("ConfigForEnv parse failed, unmarshal config failed, err=[%v]", err)
	}

	conf := configs[args.Env]
	// reset args
	conf.Env = args.Env
	if args.Port != 0 {
		conf.Port = args.Port
	}

	return conf
}
