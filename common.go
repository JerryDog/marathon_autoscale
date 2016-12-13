package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type TaskStatistics struct {
	ExecutorId string     `json:"executor_id"`
	Statistics Statistics `json:"statistics"`
}

type Statistics struct {
	CpusSystemTimeSecs float64 `json:"cpus_system_time_secs"`
	CpusUserTimeSecs   float64 `json:"cpus_user_time_secs"`
	MemRssBytes        int     `json:"mem_rss_bytes"`
	MemLimitBytes      int     `json:"mem_limit_bytes"`
	Timestamp          float64 `json:"timestamp"`
}

func check(err error) {
	if err != nil {
		log.Error(err)
		fmt.Println(err)
	}
}

func getTaskAgentStatistics(taskId string, host string) (*Statistics, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+host+":5051/monitor/statistics.json", nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var taskStatistics []TaskStatistics

	contents, err := ioutil.ReadAll(response.Body)
	//fmt.Println(string(contents))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &taskStatistics)
	if err != nil {
		return nil, err
	}

	for _, task := range taskStatistics {
		//log.Debug("taskId=" + task.Id + " running on " + task.Host)
		if task.ExecutorId == taskId {
			return &task.Statistics, nil
		}
	}
	return nil, err
}

func AppScalePoller(marathonUrl string, a appDbRow) {
	m := &Marathon{marathonUrl, a.appId, a.memPercent, a.cpuPercent,
		a.triggerCondition, a.scaleMultiplyNum, a.maxInstances, a.currentInstances,
		a.minInstances, a.overTimes}
	r, err := m.getAppDetail()
	check(err)
	sumCpuValues := 0.0
	sumMemValues := 0
	countOfTask := 0
	for k, v := range r {
		taskStats, err := getTaskAgentStatistics(k, v)
		check(err)
		// Compute CPU usage
		cpusSysTimeSecs0 := taskStats.CpusSystemTimeSecs
		cpusUserUimeSecs0 := taskStats.CpusUserTimeSecs
		timeStamp0 := taskStats.Timestamp

		time.Sleep(1 * time.Second)

		taskStats, err = getTaskAgentStatistics(k, v)
		check(err)
		cpusSysTimeSecs1 := taskStats.CpusSystemTimeSecs
		cpusUserUimeSecs1 := taskStats.CpusUserTimeSecs
		timeStamp1 := taskStats.Timestamp

		cpusTimeTotal0 := cpusSysTimeSecs0 + cpusUserUimeSecs0
		cpusTimeTotal1 := cpusSysTimeSecs1 + cpusUserUimeSecs1
		cpusTimeDelta := cpusTimeTotal1 - cpusTimeTotal0
		timestampDelta := timeStamp1 - timeStamp0

		usage := cpusTimeDelta / timestampDelta * 100.0

		// RAM usage
		memRssBytes := taskStats.MemRssBytes
		fmt.Printf("task %v mem_rss_bytes= %v\n", k, memRssBytes)
		memLimitBytes := taskStats.MemLimitBytes
		fmt.Printf("task %v mem_limit_bytes= %v\n", k, memLimitBytes)
		memUtilization := 100 * (memRssBytes / memLimitBytes)
		fmt.Printf("task %v mem Utilization= %v\n", k, memUtilization)

		sumCpuValues += usage
		sumMemValues += memUtilization
		countOfTask += 1
	}

	var appAvgCpu float64
	var appAvgMem int
	if countOfTask != 0 {
		appAvgCpu = sumCpuValues / float64(countOfTask)
		appAvgMem = sumMemValues / countOfTask
	} else {
		fmt.Printf("countOfTask == 0")
		appAvgCpu = 0.0
		appAvgMem = 0
	}
	fmt.Printf("Current Average CPU Time for app %v = %v\n", a.appId, appAvgCpu)
	fmt.Printf("Current Average Mem Utilization for app %v = %v\n", a.appId, appAvgMem)
	if a.triggerCondition == "and" {
		if appAvgCpu > float64(a.cpuPercent) && appAvgMem > a.memPercent {
			fmt.Printf("Autoscale triggered based on 'both' Mem &"+
				" CPU exceeding threshold. Over Times: %v\n", a.overTimes)
			fmt.Println("Start Autoscale...")
			m.scaleApp()
		} else {
			fmt.Println("Both values were not greater than autoscale targets.(and)")
		}
	} else {
		if appAvgCpu > float64(a.cpuPercent) || appAvgMem > a.memPercent {
			fmt.Printf("Autoscale triggered based on 'or' Mem &"+
				" CPU exceeding threshold. Over Times: %v\n", a.overTimes)
			fmt.Println("Start Autoscale...")
			m.scaleApp()
		} else {
			fmt.Println("Both values were not greater than autoscale targets.(or)")
		}
	}
	waitgroup.Done()
}
