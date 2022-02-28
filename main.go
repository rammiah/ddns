package main

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/libdns/alidns"
	"github.com/libdns/libdns"
	"github.com/tidwall/gjson"
)

const (
	DnsTypeAAAA = "AAAA"
)

var (
	AliPrd *alidns.Provider

	ErrDupStr = "The DNS record already exists"
)

func init() {
	w := log.Default().Writer()
	if o, err := os.OpenFile("ddns.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		w = io.MultiWriter(w, o)
	}
	log.SetOutput(w)
	if err := godotenv.Load(); err != nil {
		log.Printf("load .env error: %v\n", err)
		panic(err)
	}
	AliPrd = &alidns.Provider{
		AccKeyID:     os.Getenv("ACCESSKEY_ID"),
		AccKeySecret: os.Getenv("ACCESSKEY_SECRET"),
	}
}

const (
	IPv6Url = "https://ipv6.ddnspod.com"
	IPv4Url = "https://ipv4.ddnspod.com"
)

type JsonIP struct {
	IP string `json:"ip"`
}

func GetIPv6IP() net.IP {
	resp, err := http.Get(IPv6Url)
	if err != nil {
		log.Printf("get http error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read data error: %v\n", err)
		return nil
	}

	ipStr := strings.TrimSpace(string(content))

	ip := net.ParseIP(ipStr)
	if len(ip) != net.IPv6len {
		log.Printf("parse ip[%v] failed\n", ipStr)
		return nil
	}
	return ip
}

func Update(sub, dom, typ, target string) error {
	ctx := context.Background()
	var err error
	for i := 0; i < 3; i++ {
		_, err = AliPrd.SetRecords(ctx, dom, []libdns.Record{
			{
				Type:  typ,
				Name:  sub,
				Value: target,
				TTL:   10 * time.Minute,
			},
		})
		if err == nil || strings.Contains(err.Error(), ErrDupStr) {
			break
		}
	}

	if err != nil {
		log.Printf("update records error: %v\n", err)
		return err
	}
	log.Printf("update %v.%v success\n", sub, dom)
	return nil
}

func UpdateDomainsFromFile() error {
	var ip net.IP
	for i := 0; i < 3; i++ {
		ip = GetIPv6IP()
		if len(ip) == net.IPv6len {
			break
		}
	}
	if len(ip) != net.IPv6len {
		log.Printf("get ip error: nil ip\n")
		return errors.New("nil ip")
	}

	log.Printf("get ip success: %v\n", ip)
	buf, err := ioutil.ReadFile("conf/domain.json")
	if err != nil {
		log.Printf("read conf file error: %v", err)
		return err
	}

	if !gjson.ValidBytes(buf) {
		return errors.New("conf file not valid json")
	}

	js := gjson.ParseBytes(buf)
	if !js.IsArray() {
		return errors.New("no domain array detected")
	}

	domains := js.Array()
	for _, d := range domains {
		var (
			sub    = d.Get("sub").String()
			domain = d.Get("domain").String()
			typo   = d.Get("type").String()
		)
		if err := Update(sub, domain, typo, ip.String()); err != nil {
			if strings.Contains(err.Error(), ErrDupStr) {
				// duplicate, skip
				continue
			}
			return err
		}
	}

	log.Printf("update all %v domain record success\n", len(domains))
	return nil
}

func main() {
	var err error
	for i := 0; i < 3; i++ {
		if err = UpdateDomainsFromFile(); err == nil {
			log.Printf("update domain record success\n")
			return
		}
	}
	log.Printf("update domain record error: %v\n", err)
}
