package main

import (
	"flag"
	"fmt"
	"github.com/AlekSi/zabbix-sender"
	"github.com/hpcloud/tail"
	"net"
	"os"
	"regexp"
)

var (
	filename         = flag.String("file", "", "-file /path/to/file")
	regexpstring     = flag.String("regexp", "", "-regexp string want to match")
	zabbixhost       = flag.String("zabbix_host", "", "-zabbix_host zabbix host name")
	zabbixkey        = flag.String("zabbix_key", "", "-zabbix_key zabbix item key")
	zabbixserverhost = flag.String("zabbix_server_host", "", "-zabbix_server_host zabbix server host")
	zabbixserverport = flag.String("zabbix_server_port", "10051", "-zabbix_server_port zabbix server port")
	version          = flag.Bool("version", false, "-version print version")
	Version          string
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("version", Version)
		os.Exit(1)
	}

	zbx_serv_conn_str := *zabbixserverhost + ":" + *zabbixserverport

	t, err := tail.TailFile(*filename, tail.Config{
		Follow:   true,
		ReOpen:   true,
		Poll:     true,
		Location: &tail.SeekInfo{0, 2}})
	for line := range t.Lines {
		matched, err := regexp.MatchString(*regexpstring, line.Text)
		if matched == true {
			//fmt.Println(line.Time, line.Text)
			data := map[string]interface{}{*zabbixkey: line.Text}
			di := zabbix_sender.MakeDataItems(data, *zabbixhost)
			addr, _ := net.ResolveTCPAddr("tcp", zbx_serv_conn_str)
			zabbix_sender.Send(addr, di)
		}
		if nil != err {
			fmt.Println(err)
		}
	}
	if nil != err {
		fmt.Println(err)
	}
}
