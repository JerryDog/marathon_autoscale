package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
)

// main use this
type Marathon struct {
	MarathonUrl      string
	AppName          string
	MemPercent       int
	CpuPercent       int
	TriggerCondition string
	ScaleMultiplyNum float64
	MaxInstances     int
	CurrentInstances int
	MinInstances     int
	OverTimes        int
}

type marathonApp struct {
	App marathonAppDetail `json:"app"`
}

type marathonAppDetail struct {
	Id        string `json:"id"`
	Instances int    `json:"instances"`
	Tasks     []task `json:"tasks"`
}

type task struct {
	Id   string `json:"id"`
	Host string `json:"host"`
}

func (m *Marathon) getAppDetail() (map[string]string, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", m.MarathonUrl+"/v2/apps/"+m.AppName, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var appResponse marathonApp

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &appResponse)
	if err != nil {
		return nil, err
	}

	appTaskDict := make(map[string]string)

	for _, task := range appResponse.App.Tasks {
		//log.Debug("taskId=" + task.Id + " running on " + task.Host)
		appTaskDict[task.Id] = task.Host
	}

	return appTaskDict, nil
}

func (m *Marathon) scaleApp() {
	targetInstancesFloat := float64(m.CurrentInstances) * m.ScaleMultiplyNum
	targetInstances := int(math.Ceil(targetInstancesFloat))
	if int(targetInstances) > m.MaxInstances {
		log.Info("Reached the set maximum instances of %v\n", m.MaxInstances)
		fmt.Printf("Reached the set maximum instances of %v\n", m.MaxInstances)
		targetInstances = m.MaxInstances
	} else {
		targetInstances = targetInstances
	}

	fmt.Printf("Targent Instances: %v\n", targetInstances)
	str := fmt.Sprintf("%d", targetInstances)
	log.Info(m.AppName+" Targent Instances: ", str)
	var jsonStr = []byte(`{"instances": ` + str + `}`)
	req, err := http.NewRequest("PUT", m.MarathonUrl+"/v2/apps/"+m.AppName, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		log.Info("Scale_app " + m.AppName + " Success")
		log.Info("Scale_app "+m.AppName+" return status body =:", string(body))
	} else {
		log.Warning("Scale_app "+m.AppName+" Failed, status code =:", resp.Status)
		log.Warning("Scale_app "+m.AppName+" return status body =:", string(body))
	}

	fmt.Println("Scale_app return status code =:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)

	fmt.Println("Scale_app return status body =:", string(body))
}
