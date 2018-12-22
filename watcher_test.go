package discovery

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"testing"
	"time"
)

func print(event *clientv3.Event) {
	fmt.Printf("kev:%v, val:%v\n", string(event.Kv.Key), string(event.Kv.Value))
}

func TestWatch(t *testing.T) {
	watcher := NewWatcher()
	watcher.ResistHander("test-test", print)
	go watcher.run()

	for i := 0; i < 10; i++ {
		PutServerWatchKey("test-test", fmt.Sprintf("test-result-value=%v", i))
	}
	time.Sleep(3 * time.Second)
}
