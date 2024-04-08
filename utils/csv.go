package utils

import (
	"encoding/csv"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"time"
)

const (
	defaultOutput         = "result.csv"
	maxDelay              = 9999 * time.Millisecond
	minDelay              = 0 * time.Millisecond
	maxLossRate   float32 = 1.0
)

var (
	InputMaxDelay    = maxDelay
	InputMinDelay    = minDelay
	InputMaxLossRate = maxLossRate
	Output           = defaultOutput
	PrintNum         = 10
)

// 是否打印测试结果
func NoPrintResult() bool {
	return PrintNum == 0
}

// 是否输出到文件
func noOutput() bool {
	return Output == "" || Output == " "
}

type PingData struct {
	IP       *net.IPAddr
	Sended   int
	Received int
	Delay    time.Duration
}

type CloudflareIPData struct {
	*PingData
	lossRate      float32
	DownloadSpeed float64
	Country       string
}

// 计算丢包率
func (cf *CloudflareIPData) getLossRate() float32 {
	if cf.lossRate == 0 {
		pingLost := cf.Sended - cf.Received
		cf.lossRate = float32(pingLost) / float32(cf.Sended)
	}
	return cf.lossRate
}

func (cf *CloudflareIPData) toString() []string {
	result := make([]string, 6)
	result[0] = cf.IP.String()
	result[1] = strconv.Itoa(cf.Sended)
	result[2] = strconv.Itoa(cf.Received)
	result[3] = strconv.FormatFloat(float64(cf.getLossRate()), 'f', 2, 32)
	result[4] = strconv.FormatFloat(cf.Delay.Seconds()*1000, 'f', 2, 32)
	result[5] = strconv.FormatFloat(cf.DownloadSpeed/1024/1024, 'f', 2, 32)
	return result
}

func WriteFile(filename string, data []CloudflareIPData) {
	fp, err := os.Create(filename)
	if err != nil {
		log.Fatalf("创建文件[%s]失败：%v", filename, err)
		return
	}
	defer fp.Close()
	w := csv.NewWriter(fp) //创建一个新的写入文件流
	_ = w.Write([]string{"IP 地址", "已发送", "已接收", "丢包率(%)", "延迟(ms)", "下载速度(Mbps)"})
	_ = w.WriteAll(convertToString(data))
	w.Flush()
}

func ExportCsvByCountry(data []CloudflareIPData) {
	theCC := ""
	datatmp := make([]CloudflareIPData, 0)
	for i, v := range data {
		if theCC == "" {
			theCC = v.Country
		}
		if v.Country != theCC {
			filename := theCC + ".csv"
			sort.Slice(datatmp, func(i, j int) bool {
				return datatmp[i].DownloadSpeed > datatmp[j].DownloadSpeed
			})
			WriteFile(filename, datatmp)
			theCC = v.Country
			datatmp = datatmp[:0]
		}
		datatmp = append(datatmp, v)
		if i == len(data)-1 {
			filename := theCC + ".csv"
			sort.Slice(datatmp, func(i, j int) bool {
				return datatmp[i].DownloadSpeed > datatmp[j].DownloadSpeed
			})
			WriteFile(filename, datatmp)
		}
	}
}

func ExportCsv(data []CloudflareIPData) {
	if noOutput() || len(data) == 0 {
		return
	}
	fp, err := os.Create(Output)
	if err != nil {
		log.Fatalf("创建文件[%s]失败：%v", Output, err)
		return
	}
	defer fp.Close()
	gbkEncoder := simplifiedchinese.GBK.NewEncoder()
	w := csv.NewWriter(transform.NewWriter(fp, gbkEncoder))
	_ = w.Write([]string{"IP 地址", "已发送", "已接收", "丢包率", "平均延迟", "下载速度 (MB/s)"})
	_ = w.WriteAll(convertToString(data))
	w.Flush()
}

func convertToString(data []CloudflareIPData) [][]string {
	result := make([][]string, 0)
	for _, v := range data {
		result = append(result, v.toString())
	}
	return result
}

// 延迟丢包排序
type PingDelaySet []CloudflareIPData

// 延迟条件过滤
func (s PingDelaySet) FilterDelay() (data PingDelaySet) {
	if InputMaxDelay > maxDelay || InputMinDelay < minDelay { // 当输入的延迟条件不在默认范围内时，不进行过滤
		return s
	}
	if InputMaxDelay == maxDelay && InputMinDelay == minDelay { // 当输入的延迟条件为默认值时，不进行过滤
		return s
	}
	for _, v := range s {
		if v.Delay > InputMaxDelay { // 平均延迟上限，延迟大于条件最大值时，后面的数据都不满足条件，直接跳出循环
			break
		}
		if v.Delay < InputMinDelay { // 平均延迟下限，延迟小于条件最小值时，不满足条件，跳过
			continue
		}
		data = append(data, v) // 延迟满足条件时，添加到新数组中
	}
	return
}

// 丢包条件过滤
func (s PingDelaySet) FilterLossRate() (data PingDelaySet) {
	if InputMaxLossRate >= maxLossRate { // 当输入的丢包条件为默认值时，不进行过滤
		return s
	}
	for _, v := range s {
		if v.getLossRate() > InputMaxLossRate { // 丢包几率上限
			break
		}
		data = append(data, v) // 丢包率满足条件时，添加到新数组中
	}
	return
}

func (s PingDelaySet) FilterCountryV2(info *CountryInfo, countrycodeList map[string]bool) (data PingDelaySet) {
	fmt.Printf("开始查询IP的国家代号...\n")
	bfilterCountry := false
	if len(countrycodeList) > 0 {
		bfilterCountry = true
	}
	for _, v := range s {
		v.Country = info.GetCountryV2(v.IP.IP.String())
		if v.Country == "" {
			fmt.Printf("IP: %s 无法获取国家代号，已跳过\n", v.IP.IP.String())
			continue
		}
		if bfilterCountry {
			if countrycodeList[v.Country] == true {
				data = append(data, v)
			}
		} else {
			data = append(data, v)
		}
	}
	return
}

func (s PingDelaySet) FilterCountryV3(info *CountryInfo) (data PingDelaySet) {
	fmt.Printf("开始查询IP的国家代号...\n")
	for _, v := range s {
		v.Country = info.GetCountryV2(v.IP.IP.String())
		if v.Country == "" {
			data = append(data, v)
		}

	}
	return
}

func (s PingDelaySet) FilterCountryV1() (data PingDelaySet) {
	countrys := make([]IpResponse, 0)
	ips := make([]string, 0)
	querycount := 0
	fmt.Printf("开始查询IP的国家代号...\n")
	bar := NewBar(len(s), "已查:", "")
	for i, v := range s {
		if i > 0 && i%100 == 0 {
			countryCodeInfo := getCountryV1(ips)
			if countryCodeInfo == nil {
				fmt.Printf("[WARN] %s\n", "IP查询失败，请检查网络或API接口是否正常")
				continue
			}

			countrys = append(countrys, countryCodeInfo...)
			bar.Grow(100, strconv.Itoa(len(countrys)))
			ips = ips[:0]
			querycount++
			if querycount >= 15 {
				time.Sleep(61 * time.Second)
				querycount = 0
			}
		}
		ip := v.IP.IP
		ips = append(ips, ip.String())
	}

	if len(ips) > 0 {
		if querycount >= 15 {
			time.Sleep(61 * time.Second)
		}
		countryCodeInfo := getCountryV1(ips)
		countrys = append(countrys, countryCodeInfo...)
		bar.Grow(len(ips), strconv.Itoa(len(countrys)))
	}
	bar.Done()
	for i, v := range countrys {
		m := s[i]
		if v.Status == "fail" {
			m.Country = ""
		} else {
			m.Country = v.CountryCode
		}
		data = append(data, m)
	}
	return
}

func (s PingDelaySet) Len() int {
	return len(s)
}
func (s PingDelaySet) Less(i, j int) bool {
	iRate, jRate := s[i].getLossRate(), s[j].getLossRate()
	if iRate != jRate {
		return iRate < jRate
	}
	return s[i].Delay < s[j].Delay
}
func (s PingDelaySet) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// 下载速度排序
type DownloadSpeedSet []CloudflareIPData

func (s DownloadSpeedSet) Len() int {
	return len(s)
}
func (s DownloadSpeedSet) Less(i, j int) bool {
	return s[i].DownloadSpeed > s[j].DownloadSpeed
}
func (s DownloadSpeedSet) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s DownloadSpeedSet) Print() {
	if NoPrintResult() {
		return
	}
	if len(s) <= 0 { // IP数组长度(IP数量) 大于 0 时继续
		fmt.Println("\n[信息] 完整测速结果 IP 数量为 0，跳过输出结果。")
		return
	}
	dateString := convertToString(s) // 转为多维数组 [][]String
	if len(dateString) < PrintNum {  // 如果IP数组长度(IP数量) 小于  打印次数，则次数改为IP数量
		PrintNum = len(dateString)
	}
	headFormat := "%-16s%-5s%-5s%-5s%-6s%-11s\n"
	dataFormat := "%-18s%-8s%-8s%-8s%-10s%-15s\n"
	for i := 0; i < PrintNum; i++ { // 如果要输出的 IP 中包含 IPv6，那么就需要调整一下间隔
		if len(dateString[i][0]) > 15 {
			headFormat = "%-40s%-5s%-5s%-5s%-6s%-11s\n"
			dataFormat = "%-42s%-8s%-8s%-8s%-10s%-15s\n"
			break
		}
	}
	fmt.Printf(headFormat, "IP 地址", "已发送", "已接收", "丢包率", "平均延迟", "下载速度 (MB/s)")
	for i := 0; i < PrintNum; i++ {
		fmt.Printf(dataFormat, dateString[i][0], dateString[i][1], dateString[i][2], dateString[i][3], dateString[i][4], dateString[i][5])
	}
	if !noOutput() {
		fmt.Printf("\n完整测速结果已写入 %v 文件，可使用记事本/表格软件查看。\n", Output)
	}
}

func (s DownloadSpeedSet) PrintbyCountry() {
	if NoPrintResult() {
		return
	}
}
