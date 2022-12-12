package respority

/**
    @date: 2022/12/12
**/

type User struct {
	UserID uint `grom:"primarykey"`
	UserName string `grom:"unique"`
	NickName string
	PasswordDigest string
}