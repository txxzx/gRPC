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
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(workDir + "/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}// 读取配置
	  // 配置文件的文件名
	            // 配置文件的后缀
				// 获取配置文件的路径


}