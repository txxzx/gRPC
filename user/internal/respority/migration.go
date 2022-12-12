package respority

import "fmt"

/**
    @date: 2022/12/12
**/

func migration () {
	// 进行自动迁移
	err := DB.Set("gorm:tabl_options","charset=utf8").AutoMigrate(
		// 定义user结构体就可以进行自动迁移
		&User{
		},
	)
	if err !=nil {
		fmt.Errorf("migration err-> %v",err)
	}
}