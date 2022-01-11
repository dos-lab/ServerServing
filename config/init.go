package config

import (
	"flag"
	"os"
)

var conf *EachConfig

func GetConfig() *EachConfig {
	return conf
}

func InitConfig() {
	args := &args{}
	initCmdArgs(args)
	initEnvironArgs(args)
	conf = parse(args)
}

func InitConfigWithFile(path string, env ConfigurationEnv) {
	args := &args{
		ConfigPath: path,
		Env:        env,
	}
	conf = parse(args)
}

func initCmdArgs(args *args) {
	var configPath string
	var env string
	var port int
	flag.StringVar(&configPath, "config_path", "", "配置文件路径")
	flag.StringVar(&env, "env", "dev", "是否为测试测试环境，值为dev或prd")
	flag.IntVar(&port, "port", 0, "指定端口号，默认为配置文件中配置的端口号")
	flag.Parse()
	e, _ := convert2Env(env)
	if e != "" {
		args.Env = e
	}
	if port != 0 {
		args.Port = port
	}
	if configPath != "" {
		args.ConfigPath = configPath
	}
}

func initEnvironArgs(args *args) {
	configPath, ok := os.LookupEnv("CONFIG_PATH")
	if ok {
		args.ConfigPath = configPath
	}
	env, ok := os.LookupEnv("ENV")
	if ok {
		e, _ := convert2Env(env)
		if e != "" {
			args.Env = e
		}
	}
}
