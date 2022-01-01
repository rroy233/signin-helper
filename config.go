package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"signin/Logger"
)

var config *Struct

// Struct 配置文件结构体
type Struct struct {
	General struct {
		Production bool   `yaml:"production"`
		BaseUrl    string `yaml:"base_url"`
		ListenPort string `yaml:"listen_port"`
		JwtKey     string `yaml:"jwt_key"`
		AESKey     string `yaml:"AES_key"`
		AESIv      string `yaml:"AES_iv"`
	} `yaml:"genenral"`
	Db struct {
		Server    string `yaml:"server"`
		Port      string `yaml:"port"`
		Username  string `yaml:"username"`
		Password  string `yaml:"password"`
		Db        string `yaml:"db"`
		RedisAddr string `yaml:"redis_addr"`
		RedisPwd  string `yaml:"redis_pwd"`
		RedisDb   int    `yaml:"redis_db"`
	} `yaml:"db"`
	SSO struct {
		ServiceName  string `yaml:"service_name"`
		ClientId     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
	} `yaml:"sso"`
	Mail struct {
		SmtpServer      string `yaml:"smtp_server"`
		Port            string `yaml:"port"`
		TLS             bool   `yaml:"tls"`
		Username        string `yaml:"username"`
		Password        string `yaml:"password"`
		QueueBufferSize int    `yaml:"queue_buffer_size"`
	} `yaml:"mail"`
	WxPusher struct {
		AppToken string `yaml:"app_token"`
	} `yaml:"wx_pusher"`
	Cos struct {
		BucketUrl string `yaml:"bucket_url"`
		BasePath  string `yaml:"base_path"`
		SecretID  string `yaml:"secret_id"`
		SecretKey string `yaml:"secret_key"`
	} `yaml:"cos"`
}

// getConfig 读取配置文件并返回
func getConfig() {
	confFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		Logger.FATAL.Fatalln("配置文件加载失败！")
	}
	err = yaml.Unmarshal(confFile, &config)
	if err != nil {
		Logger.FATAL.Fatalln("配置文件加载失败！(" + err.Error() + ")")
	}
	return
}
