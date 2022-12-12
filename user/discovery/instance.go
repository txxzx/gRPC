package discovery

/**
    @date: 2022/12/12
**/

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc/resolver"
)

// 服务实例
type Server struct {
	Name    string `json:"name"`    // 服务名字
	Addr    string `json:"addr"`    // 服务地址
	Version string `json:"version"` // 服务版本
	Weight  int64  `json:"weight"`  // 服务权重 降级处理需要用
}

// 获取服务构建的前缀
func BuildPrefix(server Server) string {
	if server.Version == "" {
		return fmt.Sprintf("/%s/", server.Name)
	}

	return fmt.Sprintf("/%s/%s/", server.Name, server.Version)
}

// 拼接服务路径
func BuildRegisterPath(server Server) string {
	return fmt.Sprintf("%s%s", BuildPrefix(server), server.Addr)
}

// 解析服务结构信息反序列化
func ParseValue(value []byte) (Server, error) {
	server := Server{}
	if err := json.Unmarshal(value, &server); err != nil {
		return server, err
	}

	return server, nil
}

// 切割服务的路径
func SplitPath(path string) (Server, error) {
	server := Server{}

	fmt.Println(path)
	strs := strings.Split(path, "/")
	if len(strs) == 0 {
		return server, errors.New("invalid path")
	}

	server.Addr = strs[len(strs)-1]

	return server, nil
}

// Exist helper function 判断服务路径是否存在雷保中
func Exist(l []resolver.Address, addr resolver.Address) bool {
	for i := range l {
		if l[i].Addr == addr.Addr {
			return true
		}
	}

	return false
}

// Remove helper function
func Remove(s []resolver.Address, addr resolver.Address) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr.Addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}

func BuildResolverUrl(app string) string {
	return schema + ":///" + app
}
