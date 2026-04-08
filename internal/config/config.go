package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port         int    `mapstructure:"port"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	ParentFileID string `mapstructure:"parent_file_id"`
	CustomDomain string `mapstructure:"custom_domain"`
	APIToken     string `mapstructure:"api_token"`
}

var GlobalConfig Config

// InitConfig 读取并解析配置文件
func InitConfig(cfgFile string) {
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("yaml")

	// 默认值
	viper.SetDefault("port", 8080)
	viper.SetDefault("custom_domain", "")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("配置文件读取失败: %v", err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		log.Fatalf("配置解码失败: %v", err)
	}

	// 【容错防御升级】很多时候我们从网页双击复制 Secret 时会带上首尾的隐形空格或者换行符
	// 导致 123pan API 算出来的摘要完全错位并报出 “无效的登录信息”。
	// 必须在此增加系统级别的修剪操作！
	GlobalConfig.ClientID = strings.TrimSpace(GlobalConfig.ClientID)
	GlobalConfig.ClientSecret = strings.TrimSpace(GlobalConfig.ClientSecret)
	GlobalConfig.ParentFileID = strings.TrimSpace(GlobalConfig.ParentFileID)
	GlobalConfig.APIToken = strings.TrimSpace(GlobalConfig.APIToken)

	if GlobalConfig.ClientID == "" || GlobalConfig.ClientSecret == "" {
		log.Fatalf("配置错误: client_id 和 client_secret 不能为空")
	}

	if GlobalConfig.APIToken == "" {
		log.Println("【安全警告】您的 api_token 配置为空，图床接口处于裸奔危险状态！")
	}

	log.Println("成功加载配置文件，服务端口:", GlobalConfig.Port)
}
