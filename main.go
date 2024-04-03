package main

import (
	"errors"
	"log"
	"net"
	"strings"
	"time"

	"github.com/lrstanley/go-bogon"
)

var (
	ErrInvalidStatusCode  error = errors.New("did not get an acceptiable status code from the server")
	ErrFailedToDecodeBody error = errors.New("unable to decode the body")
	ErrFailedToDecodeJson error = errors.New("unexpected json format was returned")
	ErrWasNotJson         error = errors.New("response from server was not json")
	ErrDomainNotFound     error = errors.New("unable to find requested domain on cloudflare")
	ErrUnknownIpStack     error = errors.New("unknown ip stack")
)

type IpAddr struct {
	Ipv4 string
	Ipv6 string
}

func GetAddrLocal() (IpAddr, error) {
	ret := IpAddr{}
	ifaces, err := net.Interfaces()
	if err != nil {
		return IpAddr{}, err
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip := addr.(*net.IPNet).IP
			if is, _ := bogon.Is(ip.String()); is {
				continue
			}
			if strings.Contains(ip.String(), ":") {
				ret.Ipv6 = ip.String()
				continue
			}
			ret.Ipv4 = ip.String()
		}
	}
	return ret, nil
}

func main() {
	config := NewConfigClient()
	cfg := config.LoadConfig()

	if cfg.Email == "" {
		log.Println("Unable to find 'EMAIL' env value.")
		return
	}

	if cfg.Token == "" {
		log.Println("Unable to find 'API_TOKEN' env value.")
		return
	}

	if cfg.Domain == "" {
		log.Println("Unable to find 'DOMAIN' env value.")
		return
	}

	if len(cfg.Hosts) == 0 {
		log.Println("Unable to find 'HOSTS' env value.")
	}

	log.Println("Config Check: OK")

	cron := NewCron()
	log.Println("Cloudflare Check will run every second.")
	_, err := cron.scheduler.AddFunc("* * * * * *", func() {
		cron.RunCloudflareCheck(cfg)
	})
	if err != nil {
		log.Println(err)
	}
	cron.scheduler.AddFunc("* 0/1 * * * *", func() {
		cron.HelloWorldJob()
	})
	cron.scheduler.Start()

	log.Println("Application has started!")
	for {
		time.Sleep(30 * time.Second)
	}

}
