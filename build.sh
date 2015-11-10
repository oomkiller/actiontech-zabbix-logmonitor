go build -ldflags "-X main.Version='`git rev-parse --abbrev-ref HEAD`++`git rev-parse HEAD`'" actiontech-zabbix-logmonitor.go
