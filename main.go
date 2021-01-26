package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	jsip := new(JsonIP)
	if err := json.Unmarshal(content, jsip); err != nil {
		panic(err)
	} else {
		fmt.Printf("ip is: %q\n", jsip.IP)
	}
}
