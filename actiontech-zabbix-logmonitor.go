package main

import (
	"flag"
	"fmt"
	"github.com/AlekSi/zabbix-sender"
	"github.com/hpcloud/tail"
	"net"
	"net/http"
	"os"
	"regexp"
	"time"
	"encoding/json"
	"bytes"

)

var (
	filename         = flag.String("file", "", "-file /path/to/file")
	regexpstring     = flag.String("regexp", "", "-regexp string want to match")
	zabbixhost       = flag.String("zabbix_host", "", "-zabbix_host zabbix host name")
	zabbixkey        = flag.String("zabbix_key", "", "-zabbix_key zabbix item key")
	zabbixserverhost = flag.String("zabbix_server_host", "", "-zabbix_server_host zabbix server host")
	zabbixserverport = flag.String("zabbix_server_port", "10051", "-zabbix_server_port zabbix server port")
	open_falcon      = flag.Bool("open_falcon", false, "-open_falcon send to open-falcon")
	version          = flag.Bool("version", false, "-version print version")
	Version          string
)

type MetricValue struct {
	Endpoint  string      `json:"endpoint"`
	Metric    string      `json:"metric"`
	Value     int         `json:"value"`
	Step      int         `json:"step"`
	Type      string      `json:"counterType"`
	Tags      string      `json:"tags"`
	Timestamp int64       `json:"timestamp"`
}

func send2zabbix(lineText chan string){
	zbx_serv_conn_str := *zabbixserverhost + ":" + *zabbixserverport
	data := map[string]interface{}{*zabbixkey: <-lineText}
	di := zabbix_sender.MakeDataItems(data, *zabbixhost)
	addr, _ := net.ResolveTCPAddr("tcp", zbx_serv_conn_str)
	zabbix_sender.Send(addr, di)
}

func send2falcon(num  chan int){
	hostname, err := os.Hostname()
	step :=60
	ticker := time.NewTicker(time.Second * time.Duration(step)).C
	value := 0
	mvs := []*MetricValue{}		

	if err != nil {
		//log.Error("call os.Hostname() fail:", err)
	} 

        for{
		select {
			case  recv := <- num:
				value += recv
			case  <-ticker:
				mvs = append(mvs,&MetricValue{
					Endpoint:    hostname,
					Metric:      *zabbixkey,
					Value:       value,
					Step:        step,
					Type:        "GAUGE",
					Tags:        "regexpstring=" + *regexpstring,
					Timestamp:   time.Now().Unix(),
				})
				go sendData(mvs)
				value = 0
		}
	}
}

func sendData(data []*MetricValue){
	bs, err := json.Marshal(data)
	if err != nil {
		//log ..  nil, err
		return
	}
	strUrl := "http://127.0.0.1:1988/v1/push"

	_, err = http.Post(strUrl, "Content-Type: application/json", bytes.NewBuffer(bs))
	if err != nil {
		//log .. nil ,err
	}
	return

}

func main() {
	flag.Parse()


	if *version {
		fmt.Println("version", Version)
		os.Exit(1)
	}


	t, err := tail.TailFile(*filename, tail.Config{
		Follow:   true,
		ReOpen:   true,
		Poll:     true,
		Location: &tail.SeekInfo{0, 2}})
	if nil != err {
		fmt.Println(err)
		os.Exit(1)
	}

	num := make(chan int)
        if *open_falcon {
		go send2falcon(num)
	}
       
	lineText := make(chan string, 1)

	for line := range t.Lines {
		matched, err := regexp.MatchString(*regexpstring, line.Text)
		if nil != err {
			fmt.Println(err)
			os.Exit(1)
		}
		if matched == true {
			if *open_falcon { 
				 num <- 1 
			}  
			lineText <- line.Text
		 	go send2zabbix(lineText) //send2zabbix
		}
	}
}
