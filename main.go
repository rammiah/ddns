package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	_ "github.com/aliyun/alibaba-cloud-sdk-go"
)

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

func main() {
	ip := GetIPv6IP()
	if ip == nil {
		log.Printf("get ip failed\n")
	} else {
		log.Printf("get ip success: %v", ip)
	}
}
