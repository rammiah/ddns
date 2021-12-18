package main

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/libdns/alidns"
	"github.com/libdns/libdns"
)

const (
	DnsTypeAAAA = "AAAA"
)

var (
	AliPrd *alidns.Provider
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
	IPv6Url = "https://api-ipv6.ip.sb/jsonip"
	IPv4Url = "https://api-ipv4.ip.sb/jsonip"
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

	jsip := new(JsonIP)
	if err := json.Unmarshal(content, jsip); err != nil {
		log.Printf("unmarshal result [%v] error: %v\n", string(content), err)
		return nil
	}
	ip := net.ParseIP(jsip.IP)
	if ip == nil {
		log.Printf("parse ip[%v] failed\n", jsip.IP)
		return nil
	}
	return ip
}

func RunUpdate(sub, dom, typ string) {
	ip := GetIPv6IP()
	if ip == nil {
		log.Printf("get ip failed\n")
		return
	}
	log.Printf("get ip success: %v\n", ip)

	ctx := context.Background()
	_, err := AliPrd.SetRecords(ctx, dom, []libdns.Record{
		{
			Type:  typ,
			Name:  sub,
			Value: ip.String(),
			TTL:   10 * time.Minute,
		},
	})

	if err != nil {
		log.Printf("update records error: %v\n", err)
		return
	}
	log.Printf("update %v.%v success\n", sub, dom)
}

func main() {
	RunUpdate("laptop", "awsl.xin", DnsTypeAAAA)
	RunUpdate("qbit", "awsl.xin", DnsTypeAAAA)
}
