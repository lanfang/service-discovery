package discovery

import (
	"fmt"
	etcd3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/naming"
	"github.com/lanfang/go-lib/log"
	"golang.org/x/net/context"
	grpcNaming "google.golang.org/grpc/naming"
	"os"
	"os/signal"
	"syscall"
)

// Prefix should start and end with no slash
const Prefix = "Discovery/Service"

var (
	serviceKey, serviceValue string
	stopSignal               = make(chan bool, 1)
)

// Register 服务注册
func Register(name, host string, port int, prefix string) error {
	serviceValue = fmt.Sprintf("%s:%d", host, port)
	prefixTag := Prefix
	if prefix != "" {
		prefixTag = prefix
	}
	serviceKey = fmt.Sprintf("/%s/%s", prefixTag, name)
	var err error

	r := &naming.GRPCResolver{Client: client}

	log.Info("start registe server:%v", serviceKey)
	if err = r.Update(context.TODO(),
		serviceKey,
		grpcNaming.Update{Op: grpcNaming.Add, Addr: serviceValue, Metadata: "......"},
		etcd3.WithLease(nameLease)); err != nil {
		log.Error("regist %v to etcd failed :%v server exit", serviceKey, err)
		os.Exit(-1)
	}
	keepAliveOnce.Do(func() {
		go keepAlive(client, nameLease)
	})
	return err
}

func listenSignal(signals ...os.Signal) <-chan os.Signal {
	sig := make(chan os.Signal, 1)
	if len(signals) == 0 {
		signals = append(signals, os.Kill, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR2)
	}
	signal.Notify(sig, signals...)
	return sig
}

func keepAlive(c *etcd3.Client, id etcd3.LeaseID) {
	ctx, cancel := context.WithCancel(context.Background())
	ka, err := c.KeepAlive(ctx, id)
	sig := listenSignal()
	if err != nil {
		log.Error("lease keepalive err:%v, server exit", err)
		os.Exit(1)
	}
	for {
		select {
		case resp := <-ka:
			if resp != nil {
				log.Debug("keep alive leaseId:%v, ttl:%v", resp.ID, resp.TTL)
			} else {
				//ka may closed
				go keepAlive(c, id)
				return
			}
		case s := <-sig:
			log.Info("get signal [%s] keepAlive stop", s.String())
			cancel()
			c.Revoke(context.Background(), id)
			return
		}
	}
}
