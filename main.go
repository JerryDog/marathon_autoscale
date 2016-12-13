// marathon_autoscale project main.go
package main

import (
	"flag"
	"fmt"
	"os"
	//s "strings"
	"sync"
	"time"
	"runtime"
	"github.com/op/go-logging"
)

var configFilePath string
var conf Configuration
var err error
var log *logging.Logger
var waitgroup sync.WaitGroup


func init() {
	flag.StringVar(&configFilePath, "config", "marathon_autoscale.json",
		"Full path of the configuration JSON file")
	flag.Parse()
	conf, err = FromFile(configFilePath)
	check(err)
	log = getLogger(conf)
}

func main() {
	
        runtime.GOMAXPROCS(runtime.NumCPU())
	for {
		fmt.Println("Start to check apps")
		appList := GetAppList(conf)
		fmt.Println(appList)

		for _, app := range appList.Rows {
			waitgroup.Add(1)
			go AppScalePoller(conf.Marathon.MarathonUrl, app)
		}

		waitgroup.Wait()

		fmt.Println("sleep 30s...\n\n\n\n")
		time.Sleep(30 * time.Second)
	}
	//log.Info("App list need to check -->" + s.Join(appList, ","))

}

func getLogger(conf Configuration) *logging.Logger {
	var log = logging.MustGetLogger("example")

	var format = logging.MustStringFormatter(
		`%{color}%{time:2006-01-02 15:04:05.000} %{shortfunc} > %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	logFile, err := os.OpenFile(conf.Marathon.LogPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModeType)
	if err != nil {
		fmt.Println(err)
	}
	backend := logging.NewLogBackend(logFile, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)
	return log
}
