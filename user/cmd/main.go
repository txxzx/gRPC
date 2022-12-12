package main

/**
    @date: 2022/12/12
**/

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/txxzx/gRPC/user/internal/service"

	"google.golang.org/grpc"
	"net"
	"github.com/txxzx/gRPC/user/user/config"
	"github.com/txxzx/gRPC/user/user/discovery"
	"github.com/txxzx/gRPC/user/internal/respority"
)

// 启动入口
func main() {
	// 初始化配置文件
	config.InitConfig()
	// 初始化数据
	respority.InitDB()
	// 取出etcd 地址 服务注册到etcd里面
	etcdAddress := []string{viper.GetString("etcd.address")}
	// 服务注册
	etcdRegister := discovery.NewRegister(etcdAddress, logrus.New())
	grpcAddress := viper.GetString("server.grpcAddress")
	defer etcdRegister.Stop()
	userNode := discovery.Server{
		Name: viper.GetString("server.domain"),
		Addr: grpcAddress,
	}
	server := grpc.NewServer()
	defer server.Stop()
	// 绑定service服务

	service.RegisterUserServiceServer(server, handler.NewUserService())
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		panic(err)
	}
	if _, err := etcdRegister.Register(userNode, 10); err != nil {
		panic(fmt.Sprintf("start server failed, err: %v", err))
	}
	logrus.Info("server started listen on ", grpcAddress)
	// 对这个服务进行监听
	if err := server.Serve(lis); err != nil {
		panic(err)
	}
}