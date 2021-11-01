package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/joho/godotenv"
)

var (
	dnsCli *alidns.Client
)

func init() {
	w := log.Default().Writer()
	if o, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		w = io.MultiWriter(w, o)
	}
	log.SetOutput(w)
	if err := godotenv.Load(); err != nil {
		log.Printf("load .env error: %v\n", err)
		panic(err)
	}

	cli, err := alidns.NewClientWithAccessKey("cn-beijing", os.Getenv("ACCESSKEY_ID"), os.Getenv("ACCESSKEY_SECRET"))
	if err != nil {
		log.Printf("create client error: %v\n", err)
		panic(err)
	}
	dnsCli = cli
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

func QueryRecordID(name string) (id, value string, _ error) {
	req := alidns.CreateDescribeSubDomainRecordsRequest()
	req.Scheme = "https"

	req.SubDomain = name

	resp, err := dnsCli.DescribeSubDomainRecords(req)
	if err != nil {
		log.Printf("query sub domain records error: %v\n", err)
		return "", "", err
	}
	for _, rec := range resp.DomainRecords.Record {
		if rec.Type == "AAAA" {
			log.Printf("query record success: %v, ip %v\n", rec.RecordId, rec.Value)
			return rec.RecordId, rec.Value, nil
		}
	}

	return "", "", errors.New("no results")
}

func UpdateDdns(name string, ip net.IP) error {
	recId, val, err := QueryRecordID(name)
	if err != nil {
		log.Printf("query record id error: %v\n", err)
		return err
	}
	log.Printf("query name %v success, ip %v, record id %v\n", name, val, recId)

	if val == ip.String() {
		log.Printf("ip is same, input[%v], exists[%v], return success\n", ip.String(), val)
		return nil
	}

	req := alidns.CreateUpdateDomainRecordRequest()
	req.Scheme = "https"
	req.RR = name[:strings.Index(name, ".")]
	req.RecordId = recId
	req.Type = "AAAA"
	req.Value = ip.String()

	resp, err := dnsCli.UpdateDomainRecord(req)
	if err != nil {
		log.Printf("update record error: %v\n", err)
		return err
	}

	log.Printf("request response: success %v, request id %v\n", resp.IsSuccess(), resp.RequestId)

	return nil
}

func RunUpdate() {
	ip := GetIPv6IP()
	if ip == nil {
		log.Printf("get ip failed\n")
		return
	}
	log.Printf("query ip success: %v\n", ip)

	name := "laptop.awsl.xin"

	log.Printf("get ip success: %v", ip)
	if err := UpdateDdns(name, ip); err != nil {
		log.Printf("update ip error: %v\n", err)
		return
	}

	log.Printf("update success")
}

func main() {
	RunUpdate()
}
