package discovery

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Service struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Weight   int    `json:"weight,omitempty"`
	Text     string `json:"text,omitempty"`
}

func (s *Service) ToString() string {
	x, _ := json.Marshal(s)
	return string(x)
}

type Domain struct {
	Name string
	Srv  Service
}

const (
	TTL          = 9 //seconds
	domainPrefix = "/skydns/com/service/in"
)

var (
	domain        Domain
	keepAliveOnce sync.Once
)

func RegistDomain(server, host, port string /*:port*/) string {
	return domainRegist(server, host, port, GetMD5Hash(hostname), true)
}
func UnRegistDomain() {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	client.Delete(ctx, domain.Name)
	cancel()
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func domainRegist(server, host, port, identifier string, timeout bool) string {
	if server == "" {
		server = filepath.Base(os.Args[0])
	}
	domain = Domain{
		Name: genDomain(server, identifier),
	}
	var err error
	if domain.Srv, err = genServices(host, port); err != nil {
		log.Fatalf("get domain service err:%v exit...", err)
	}
	opts := make([]clientv3.OpOption, 0)
	if timeout {
		opts = append(opts, clientv3.WithLease(nameLease))
	}

	if _, err := client.Put(context.TODO(),
		domain.Name,
		domain.Srv.ToString(),
		opts...); err != nil {
		log.Fatalf("regist domain %+v, err:%v", domain, err)
	}
	keepAliveOnce.Do(func() {
		go keepAlive(client, nameLease)
	})
	return FormatName(domain.Name)
}

func genDomain(server, id string) string {
	domain := fmt.Sprintf("%v/%v", domainPrefix, server)
	if id != "" {
		domain = fmt.Sprintf("%v/unique_%v", domain, id)
	}
	return domain
}

func genServices(host, port string) (Service, error) {
	var retErr error
	var srv Service
	for loop := true; loop; loop = false {
		if host != "" {
			break
		}
		addrSlice, err := net.InterfaceAddrs()
		if nil != err {
			retErr = err
			log.Printf("Get local IP addr failed!!!")
			break
		}
		for _, addr := range addrSlice {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if nil != ipnet.IP.To4() {
					host = ipnet.IP.String()
				}
				if strings.HasPrefix(host, "172.") || strings.HasPrefix(host, "192.") || strings.HasPrefix(host, "10.") {
					break
				}
			}
		}
	}
	srv = Service{
		Host:     host,
		Priority: 10,
		Weight:   10,
		Text:     fmt.Sprintf("domain for %v", filepath.Base(os.Args[0])),
	}
	if port != "" {
		if p, err := strconv.Atoi(strings.Trim(port, ":")); err == nil && p > 0 {
			srv.Port = p
		}
	}
	return srv, retErr
}

func FormatName(s string) string {
	l := strings.Split(s, "/")
	if len(l) == 0 {
		return ""
	}
	for i, j := 1, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
	return strings.Join(l[2:len(l)-1], ".")
}
