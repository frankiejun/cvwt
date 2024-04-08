package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/oschwald/maxminddb-golang"
	"io"
	"net"
	"net/http"
)

type IpResponse struct {
	Status      string `json:"status"`
	CountryCode string `json:"countryCode"`
}

// 每分钟15次查询，每次最多100个ip
func getCountryV1(ips []string) (coutryCodes []IpResponse) {
	jsonData, err := json.Marshal(ips)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	resp, err := http.Post("http://ip-api.com/batch", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	data := make([]IpResponse, 0)
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for _, v := range data {
		coutryCodes = append(coutryCodes, v)
	}

	return
}

// 使用本地mmdb文件查国家代码
type Country struct {
	GeoNameID uint              `maxminddb:"geoname_id"`
	IsoCode   string            `maxminddb:"iso_code"`
	Names     map[string]string `maxminddb:"names"`
}

type Record struct {
	C Country `maxminddb:"country"`
}

type CountryInfo struct {
	db *maxminddb.Reader
}

func NewCountryInfo(db *maxminddb.Reader) *CountryInfo {
	return &CountryInfo{
		db: db,
	}
}

func (c *CountryInfo) GetCountryV2(ip string) string {

	ipinfo := net.ParseIP(ip)
	var record Record
	err := c.db.Lookup(ipinfo, &record)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return record.C.IsoCode
}

func (c *CountryInfo) Done() {
	defer c.db.Close()
}
