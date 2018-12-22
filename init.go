package discovery

import (
	"github.com/lanfang/go-lib/log"
	"github.com/satori/go.uuid"
	"os"
	"path/filepath"
	"sync"
)

const (
	etcdENV = "ETCD_ADDR"
)

var (
	newClientOne sync.Once
	hostname     string
)

func init() {
	name, err := os.Hostname()
	if err != nil {
		log.Error("get hostname err:%v, server exit", err)
		uid := uuid.NewV4()
		name = uid.String()
	}
	hostname = name
	etcdAddr := os.Getenv(etcdENV)
	newClientOne.Do(func() {
		newEtcdClient(etcdAddr)
	})
	RegistDomain(filepath.Base(os.Args[0]), "", "")
	watch = NewWatcher()
	watch.ResistHander(Loglevel, onLoglevelChanged)
	go watch.run()
}
