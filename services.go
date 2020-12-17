package main

import (
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

func getSvcPorts(result gjson.Result) string {
	var ports string
	result.ForEach(func(key, value gjson.Result) bool {
		vars := gjson.GetMany(value.String(), `protocol`, `port`)
		ports = ports + vars[0].String() + "/" + vars[1].String() + " "
		return true
	})
	return ports
}

func getServices() {
	var csvHeader = []string{"Cluster", "Namespace", "Name", "Type", "ClusterIP", "ExternalIP", "Ports", "Affinity", "CreationDate", "Version", "UID"}
	var csvData []string
	var startTime time.Time
	var duration time.Duration
	var xR int = 1
	var sheetName string = "Services"
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

		apiurl = cfg.Clusters[i].BaseURL + "/api/v1/services?limit=1000"
		body, _ := getRest(apiurl, cfg.Clusters[i].Token)

		// Loop in list items and export data into Excel file
		items := gjson.GetBytes(body, "items")
		items.ForEach(func(key, value gjson.Result) bool {
			vars := gjson.GetMany(value.String(),
				`metadata.namespace`, `metadata.name`, `spec.type`, `spec.clusterIP`, `spec.externalIPs.0`,
				`spec.ports`, `spec.sessionAffinity`, `metadata.creationTimestamp`, `metadata.resourceVersion`, `metadata.uid`)

			csvData = nil
			csvData = append(csvData, cfg.Clusters[i].Name)         // Cluster Name
			csvData = append(csvData, vars[0].String())             // Namespace Name
			csvData = append(csvData, vars[1].String())             // Service Name
			csvData = append(csvData, vars[2].String())             // Type
			csvData = append(csvData, vars[3].String())             // ClusterIP
			csvData = append(csvData, vars[4].String())             // ExternalIP
			csvData = append(csvData, getSvcPorts(vars[5]))         // Ports
			csvData = append(csvData, vars[6].String())             // Session Affinity
			csvData = append(csvData, formatDate(vars[7].String())) // Creation Timestamp
			csvData = append(csvData, vars[8].String())             // Resource Version
			csvData = append(csvData, vars[9].String())             // UID

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
