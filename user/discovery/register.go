package discovery

/**
   @date: 2022/12/12
**/

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// 定义一个注册实例
type Register struct {
	// 地址
	EtcdAddrs   []string
	// 超时时间
	DialTimeout int
	// 是否关闭
	closeCh     chan struct{}
	// 租约
	leasesID    clientv3.LeaseID
	// 心跳检验的保护，是否活着
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse
	// 服务信息
	srvInfo Server
	// 时间
	srvTTL  int64
	// 客户端
	cli     *clientv3.Client
	// 日志
	logger  *logrus.Logger
}

// 新建一个实例

// NewRegister create a register based on etcd
func NewRegister(etcdAddrs []string,logger *logrus.Logger) *Register {
	return &Register{
		EtcdAddrs:   etcdAddrs,
		DialTimeout: 3,
		 logger:      logger,
	}
}

// Register a service  基于Register的对象注册服务 初始化自己的实例
func (r *Register) Register(srvInfo Server, ttl int64) (chan<- struct{}, error) {
	var err error

	// 对地址进行切割
	if strings.Split(srvInfo.Addr, ":")[0] == "" {
		return nil, errors.New("invalid ip address")
	}

	// 配置客户端
	if r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	}); err != nil {
		return nil, err
	}

	r.srvInfo = srvInfo
	r.srvTTL = ttl

	if err = r.register(); err != nil {
		return nil, err
	}

	// make一个关闭的channel
	r.closeCh = make(chan struct{})

	// 服务节点是可靠的高可活
	go r.keepAlive()

	return r.closeCh, nil
}

 // 新建etcd自带的实例
func (r *Register) register() error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.DialTimeout)*time.Second)
	defer cancel()

	// 定义一个新的租约
	leaseResp, err := r.cli.Grant(ctx, r.srvTTL)
	if err != nil {
		return err
	}
	 // 租约id传进去
	r.leasesID = leaseResp.ID

	if r.keepAliveCh, err = r.cli.KeepAlive(context.Background(), r.leasesID); err != nil {
		return err
	}

	data, err := json.Marshal(r.srvInfo)
	if err != nil {
		return err
	}

	// 将服务push 到服务注册那里面
	_, err = r.cli.Put(context.Background(),
		BuildRegisterPath(r.srvInfo), string(data),
		clientv3.WithLease(r.leasesID))

	return err
}

// Stop stop register
func (r *Register) Stop() {
	r.closeCh <- struct{}{}
}

// unregister 删除节点。删除服务
func (r *Register) unregister() error {
	// 上下文context
	_, err := r.cli.Delete(context.Background(), BuildRegisterPath(r.srvInfo))
	return err
}

func (r *Register) keepAlive() {
	ticker := time.NewTicker(time.Duration(r.srvTTL) * time.Second)

	for {
		select {
		// 检测有没有关闭，关闭的了就删除服务
		case <-r.closeCh:
			if err := r.unregister(); err != nil {
				 r.logger.Error("unregister failed, error: ", err)
			}
			// 是否关闭成功
			if _, err := r.cli.Revoke(context.Background(), r.leasesID); err != nil {
				r.logger.Error("revoke failed, error: ", err)
			}
			// 如果没有存活就进行注册
		case res := <-r.keepAliveCh:
			if res == nil {
				// 注册一下
				if err := r.register(); err != nil {
					r.logger.Error("register failed, error: ", err)
				}
			}
			// 超时器 超时进行注册
		case <-ticker.C:
			if r.keepAliveCh == nil {
				if err := r.register(); err != nil {
					r.logger.Error("register failed, error: ", err)
				}
			}
		}
	}
}

func (r *Register) UpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		weightstr := req.URL.Query().Get("weight")
		weight, err := strconv.Atoi(weightstr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		var update = func() error {
			r.srvInfo.Weight = int64(weight)
			data, err := json.Marshal(r.srvInfo)
			if err != nil {
				return err
			}

			_, err = r.cli.Put(context.Background(), BuildRegisterPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID))
			return err
		}

		if err := update(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		_, _ = w.Write([]byte("update server weight success"))
	}
}

func (r *Register) GetServerInfo() (Server, error) {
	resp, err := r.cli.Get(context.Background(), BuildRegisterPath(r.srvInfo))
	if err != nil {
		return r.srvInfo, err
	}

	server := Server{}
	if resp.Count >= 1 {
		if err := json.Unmarshal(resp.Kvs[0].Value, &server); err != nil {
			return server, err
		}
	}

	return server, err
}
