package main

import (
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

func getDaemonsets() {
	var csvHeader = []string{"Cluster", "Namespace", "Name", "Desired", "Current", "Ready", "UpToDate", "Available", "Misscheduled", "NodeSelector",
		"UpdateStrategy", "CreationTime", "Version", "UID"}
	var csvData []interface{}
	var startTime time.Time
	var duration time.Duration
	var xR int = 1
	var sheetName string = "DaemonSets"
	var apiurl string

	// Initialize Excel sheet
	index := xf.NewSheet(sheetName)
	xf.SetActiveSheet(index)
	xf.SetSheetRow(sheetName, "A1", &csvHeader)

	info.Printf("%s: Section started\n", sheetName)
	startTime = time.Now()

	for i := 0; i < len(cfg.Clusters); i++ {
		if !cfg.Clusters[i].Enable {
			continue
		}
		boolCheck, msg := checkClusterAPI(cfg.Clusters[i].BaseURL, cfg.Clusters[i].Token)
		if !boolCheck {
			erro.Printf("%s: %s: %s\n", sheetName, cfg.Clusters[i].Name, msg)
			continue
		}
		info.Printf("%s: Working on %s\n", sheetName, cfg.Clusters[i].Name)

		apiurl = cfg.Clusters[i].BaseURL + "/apis/apps/v1/daemonsets?limit=1000"
		body, _ := getRest(apiurl, cfg.Clusters[i].Token)

		// Loop in list items and export data into Excel file
		items := gjson.GetBytes(body, "items")
		items.ForEach(func(key, value gjson.Result) bool {
			vars := gjson.GetMany(value.String(),
				`metadata.namespace`, `metadata.name`, `status.desiredNumberScheduled`, `status.currentNumberScheduled`, `status.numberReady`,
				`status.updatedNumberScheduled`, `status.numberAvailable`, `status.numberMisscheduled`,
				`spec.template.spec.nodeSelector`, `spec.updateStrategy.type`,
				`metadata.creationTimestamp`, `metadata.resourceVersion`, `metadata.uid`)

			csvData = nil
			csvData = append(csvData, cfg.Clusters[i].Name) // Cluster Name
			csvData = append(csvData, vars[0].String())     // Namespace Name
			csvData = append(csvData, vars[1].String())     // DaemonSet Name
			csvData = append(csvData, vars[2].String())     // DaemonSet Desired Scheduled
			csvData = append(csvData, vars[3].String())     // DaemonSet Currently Scheduled
			csvData = append(csvData, vars[4].String())     // DaemonSet Ready Scheduled
			csvData = append(csvData, vars[5].String())     // DaemonSet Updated Scheduled
			csvData = append(csvData, vars[6].String())     // DaemonSet Available Scheduled
			csvData = append(csvData, vars[7].String())     // DaemonSet Misscheduled
			csvData = append(csvData, vars[8].String())     // Node Selector(s)
			csvData = append(csvData, vars[9].String())     // Update Strategy Type
			csvData = append(csvData, vars[10].String())    // Creation Timestamp
			csvData = append(csvData, vars[11].String())    // Resource Version
			csvData = append(csvData, vars[12].String())    // UID

			xR++
			cell, _ := excelize.CoordinatesToCellName(1, xR)
			xf.SetSheetRow(sheetName, cell, &csvData)
			return true
		})
	}
	formatTable(sheetName, len(csvHeader))

	duration = time.Since(startTime)
	info.Printf("%s: Section ended in %s\n", sheetName, duration)
}
