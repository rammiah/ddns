package main

import (
	"io/ioutil"
	"net/http"

	_ "github.com/aliyun/alibaba-cloud-sdk-go"
	"github.com/bitly/go-simplejson"
)

const (
	IPv6Url = "https://api-ipv6.ip.sb/jsonip"
	IPv4Url = "https://api-ipv4.ip.sb/jsonip"
)

func main() {
	resp, err := http.Get(IPv4Url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	simplejson.NewJson(content)
}
