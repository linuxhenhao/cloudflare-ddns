package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/robfig/cron/v3"
)

type cronClient struct {
	scheduler *cron.Cron
	cfMap     map[string]*DnsDetails
}

var (
	ErrNoPublicIp = errors.New("no public ip addr")
)

func NewCron() cronClient {
	c := cronClient{
		scheduler: cron.New(cron.WithSeconds()),
		cfMap:     make(map[string]*DnsDetails),
	}

	return c
}

func checkStackAndAddr(ipa IpAddr, cfg ConfigModel) error {
	if cfg.IpStack == "ipv6" && ipa.Ipv6 == "" {
		return fmt.Errorf("ipv6: %w", ErrNoPublicIp)
	}
	if cfg.IpStack == "ipv4" && ipa.Ipv4 == "" {
		return fmt.Errorf("ipv4: %w", ErrNoPublicIp)
	}
	if cfg.IpStack == "dual" && (ipa.Ipv4 == "" || ipa.Ipv6 == "") {
		return fmt.Errorf("dual mode: %w", ErrNoPublicIp)
	}
	return nil
}

func (c cronClient) RunCloudflareCheck(cfg ConfigModel) {
	log.Println("Starting check...")
	log.Println("Checking the current IP Address")
	currentIp, err := GetAddrLocal()
	if err != nil {
		log.Println(err)
		return
	}
	if err := checkStackAndAddr(currentIp, cfg); err != nil {
		log.Println(err)
		return
	}

	cf := NewCloudFlareClient(cfg.Token, cfg.Email)
	log.Println("Checking domain information on CloudFlare")
	domainDetails, err := cf.GetDomainByName(cfg.Domain)
	if err != nil {
		log.Println("Unable to get information from CloudFlare.")
		log.Println("Double check the API Token to make sure it's valid.")
		log.Println(err)
		return
	}

	for _, host := range cfg.Hosts {
		hostname := fmt.Sprintf("%v.%v", host, cfg.Domain)
		log.Printf("Reviewing '%v'", hostname)
		dns, exist := c.cfMap[hostname]
		if !exist {
			dns, err = cf.GetDnsEntriesByDomain(domainDetails.Result[0].ID, host, cfg.Domain)
			if err != nil {
				log.Println("failed to collect dns entry")
				return
			}
			c.cfMap[hostname] = dns
		}
		updated := false
		switch cfg.IpStack {
		case "ipv4":
			updated = update(cf, currentIp.Ipv4, "A", dns)
		case "ipv6":
			updated = update(cf, currentIp.Ipv6, "AAAA", dns)
		case "dual":
			updated = update(cf, currentIp.Ipv4, "A", dns) || update(cf, currentIp.Ipv6, "AAAA", dns)
		default:
			log.Printf("unknown ipstack=%s\n", cfg.IpStack)
			return
		}
		if updated {
			delete(c.cfMap, hostname)
		}
	}
	log.Println("Done!")
}

func update(cf *CloudFlareClient, ip, t string, dns *DnsDetails) bool {
	for _, item := range dns.Result {
		if item.Type == t && item.Content != ip {
			log.Printf("IP Address no longer matches, sending an update, from %s to %s\n", item.Content, ip)
			err := cf.UpdateDnsEntry(item, ip)
			if err != nil {
				log.Printf("Failed to update the DNS record to %s!\n", ip)
			}
			return true
		}
	}
	log.Printf("no update for ip=%s", ip)
	return false
}

func (c cronClient) HelloWorldJob() {
	log.Print("Hello World")
}
