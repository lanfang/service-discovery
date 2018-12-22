package discovery

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/lanfang/go-lib/log"
	"os"
	"path/filepath"
	"time"
)

//watch config changed
var (
	watchPrefix = fmt.Sprintf("/conf/watch/%v", filepath.Base(os.Args[0]))
	Loglevel    = "loglevel"
	watch       *Watcher
)

type WatcherHandler func(event *clientv3.Event)

func ResistEventHander(event string, handler WatcherHandler) {
	watch.ResistHander(event, handler)
}

func genServerWatchKey(key string) string {
	return fmt.Sprintf("%v/%v", watchPrefix, key)
}

func NewWatcher() *Watcher {
	watch := &Watcher{
		client:        client,
		evnentHandler: make(map[string]WatcherHandler),
	}
	watch.ctx, watch.cancel = context.WithCancel(context.Background())
	return watch
}

type Watcher struct {
	client        *clientv3.Client
	evnentHandler map[string]WatcherHandler
	ctx           context.Context
	cancel        context.CancelFunc
}

func (w *Watcher) ResistHander(key string, handler WatcherHandler) {
	w.evnentHandler[genServerWatchKey(key)] = handler
}

func (w *Watcher) run() {
	var eventChan clientv3.WatchChan
	eventChan = client.Watch(w.ctx, watchPrefix, clientv3.WithPrefix())
LOOP:
	for {
		select {
		case resp := <-eventChan:
			if err := resp.Err(); err != nil {
				log.Error("watch err %v", err)
				time.Sleep(5 * time.Second)
				eventChan = client.Watch(w.ctx, watchPrefix, clientv3.WithPrefix())
			} else {
				for _, e := range resp.Events {
					key := string(e.Kv.Key)
					if h, ok := w.evnentHandler[key]; ok {
						h(e)
					}
				}
			}
		case <-w.ctx.Done():
			client.Watcher.Close()
			break LOOP
		}
	}
}

func (w *Watcher) Close() {
	w.cancel()
}

func onLoglevelChanged(event *clientv3.Event) {
	switch event.Type {
	case clientv3.EventTypePut:
		//do something
		break
	case clientv3.EventTypeDelete:
		//do nothing
	}
}

func PutServerWatchKey(key, val string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	_, err := client.Put(ctx, genServerWatchKey(key), val)
	if err != nil {
		log.Error("put etcd key %v, value %v, err %v", key, val, err)
	}
	cancel()
	return err
}
