package discovery

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"log"
	"strings"
	"time"
)

var (
	client      *clientv3.Client
	nameLease   clientv3.LeaseID
	defaultAddr = "http://127.0.0.1:2379"
)

func GetClient() *clientv3.Client {
	return client
}

func newEtcdClient(addr string) {
	if addr == "" {
		addr = defaultAddr
	}
	var err error
	client, err = clientv3.New(clientv3.Config{
		Endpoints: strings.Split(addr, ","),
	})
	if err != nil {
		log.Fatalf("init etcd client with %v, err %v, server exit", defaultAddr, err)
	}
	newLease()
}

func newLease() {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	resp, err := client.Grant(ctx, int64(TTL))
	if err != nil {
		log.Fatalf("etcd grant lease failed url:%+v, err:%v, server exit", client.Endpoints(), err)
	}
	nameLease = resp.ID
	cancel()
}
