package config

import (
	"os"

	"github.com/spf13/viper"
)

/**
    @date: 2022/12/7
**/

func InitConfig(){
	// 获取当前的工作目录
	workDir,_ := os.Getwd()  // 配资本金文件的路径
	viper.SetConfigName("config") // 配置文件的文件名
	viper.SetConfigType("yml")// 配置文件的后缀
	viper.AddConfigPath(workDir + "/config") // 获取配置文件的路径
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}// 读取配置





}