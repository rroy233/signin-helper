package main

import (
	"io/ioutil"

	Logger "github.com/rroy233/logger"
	"gopkg.in/yaml.v2"
)

var config *Struct

// Struct 配置文件结构体
type Struct struct {
	General struct {
		Production   bool   `yaml:"production"`
		BaseUrl      string `yaml:"base_url"`
		ListenPort   string `yaml:"listen_port"`
		JwtKey       string `yaml:"jwt_key"`
		AESKey       string `yaml:"AES_key"`
		AESIv        string `yaml:"AES_iv"`
		RandomPicAPI string `yaml:"random_pic_api"`
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
	StatisticsReport struct {
		V651La    bool   `yaml:"v6_51_la"`
		V651LaJs1 string `yaml:"v6_51_la_js1"`
		V651LaJs2 string `yaml:"v6_51_la_js2"`
	} `yaml:"statistics_report"`
	Logger struct {
		Stdout               bool   `yaml:"stdout"`
		StoreFile            bool   `yaml:"storeFile"`
		RemoteReport         bool   `yaml:"remoteReport"`
		RemoteReportUrl      string `yaml:"remoteReportUrl"`
		RemoteReportQueryKey string `yaml:"remoteReportQueryKey"`
	}
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
