package config

import (
	"log"
	"os"
	"path/filepath"
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

	// 映射环境变量并设置默认值防穿透，让用户能够通过 Docker ENV 变量纯净启动！
	viper.AutomaticEnv()
	viper.SetDefault("port", 8080)
	viper.SetDefault("custom_domain", "")
	viper.SetDefault("client_id", "")
	viper.SetDefault("client_secret", "")
	viper.SetDefault("parent_file_id", "")
	viper.SetDefault("api_token", "")

	err := viper.ReadInConfig()

	// 【部署容错机制】：当没有配置文件，且！系统环境变量中也没有提取到 client_id 等有效配置时，才触发布署断绝。
	if err != nil && viper.GetString("client_id") == "" {
		if _, statErr := os.Stat(cfgFile); os.IsNotExist(statErr) {
			log.Printf("⚠️ 尚未检测到物理配置节点及外部环境变量，系统正在初始化框架包至：[%s]", cfgFile)
			if err := os.MkdirAll(filepath.Dir(cfgFile), 0755); err != nil {
				log.Fatalf("❌ 构建部署文件夹遇到致命阻塞: %v", err)
			}

			defaultYaml := `# 123pan 私人图床配置文件

# 服务器端口
port: 8080

# 123云盘开放平台应用参数 (您必须去开放控制台申请并填入)
client_id: ""
client_secret: ""

# 123云盘专属存放目录的指定 ID (如果打算传到根目录则保留 "" 留空)
parent_file_id: ""

# 您的专属大门密匙！暴露公网请千万修改它防御盗刷
api_token: "PRIVATE_123_KEY"
`
			if err := os.WriteFile(cfgFile, []byte(defaultYaml), 0644); err != nil {
				log.Fatalf("❌ 框架配置文件剥离解压失败: %v", err)
			}

			// 主动断开并友善通知管理员
			log.Fatalf("✅ 底层初始部署圆满完成！你可以打开刚生成的 %s 填写密钥，或者直接通过在 Docker/系统环境变量中注入 CLIENT_ID 等参数来免密运行！", cfgFile)
		}
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
