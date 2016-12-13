package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type appDbRows struct {
	Rows []appDbRow
}

type appDbRow struct {
	appId            string
	memPercent       int
	cpuPercent       int
	triggerCondition string
	scaleMultiplyNum float64
	maxInstances     int
	currentInstances int
	minInstances     int
	overTimes        int
}

func (r *appDbRows) AddItem(row appDbRow) []appDbRow {
	r.Rows = append(r.Rows, row)
	return r.Rows
}

func GetAppList(conf Configuration) *appDbRows {
	db, err := sql.Open("mysql", conf.Marathon.DBUser+":"+
		conf.Marathon.DBPass+"@tcp("+
		conf.Marathon.DBHost+":"+
		conf.Marathon.DBPort+")/"+
		conf.Marathon.DBName+"?charset=utf8")
	check(err)

	defer db.Close()
	// 获取有自动伸缩功能的 app 列表
	rows, err := db.Query("SELECT app_id,mem_percent,cpu_percent," +
		"trigger_condition,scale_multiply_num,max_instances,instances_num," +
		"min_instances,over_times FROM apps_manage_appinfo where auto_scale=1")
	check(err)
	appList := &appDbRows{}
	for rows.Next() {
		var appId string
		var memPercent int
		var cpuPercent int
		var triggerCondition string
		var scaleMultiplyNum float64
		var maxInstances int
		var currentInstances int
		var minInstances int
		var overTimes int
		err = rows.Scan(&appId, &memPercent, &cpuPercent, &triggerCondition,
			&scaleMultiplyNum, &maxInstances, &currentInstances,
			&minInstances, &overTimes)
		check(err)
		dbRow := &appDbRow{appId, memPercent, cpuPercent, triggerCondition,
			scaleMultiplyNum, maxInstances, currentInstances,
			minInstances, overTimes}
		appList.AddItem(*dbRow)
	}
	return appList
}
