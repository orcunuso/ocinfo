package main

import (
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

func getRoutes() {
	var csvHeader = []string{"Cluster", "Namespace", "Name", "Hostname", "Path", "Target", "TargetPort", "WildcardPolicy", "TLS.Termination",
		"TLS.EdgeTermination", "CreationTime", "Version", "UID"}
	var csvData []string
	var startTime time.Time
	var duration time.Duration
	var xR int = 1
	var sheetName string = "Routes"
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

		apiurl = cfg.Clusters[i].BaseURL + "/apis/route.openshift.io/v1/routes?limit=1000"
		body, _ := getRest(apiurl, cfg.Clusters[i].Token)

		// Loop in list items and export data into Excel file
		items := gjson.GetBytes(body, "items")
		items.ForEach(func(key, value gjson.Result) bool {
			vars := gjson.GetMany(value.String(),
				`metadata.namespace`, `metadata.name`, `spec.host`, `spec.path`, `spec.to.kind`, `spec.to.name`, `spec.to.weight`, `spec.port.targetPort`,
				`spec.wildcardPolicy`, `spec.tls.termination`, `spec.tls.insecureEdgeTerminationPolicy`,
				`metadata.creationTimestamp`, `metadata.resourceVersion`, `metadata.uid`)

			csvData = nil
			csvData = append(csvData, cfg.Clusters[i].Name)                                           // Cluster Name
			csvData = append(csvData, vars[0].String())                                               // Namespace Name
			csvData = append(csvData, vars[1].String())                                               // Route Name
			csvData = append(csvData, vars[2].String())                                               // Route FQDN
			csvData = append(csvData, vars[3].String())                                               // Route Path
			csvData = append(csvData, vars[4].String()+"/"+vars[5].String()+"("+vars[6].String()+")") // Route Target
			csvData = append(csvData, vars[7].String())                                               // Route Target Port
			csvData = append(csvData, vars[8].String())                                               // Wildcard Policy
			csvData = append(csvData, vars[9].String())                                               // TLS Termination
			csvData = append(csvData, vars[10].String())                                              // TLS Insecure Edge Termination Policy
			csvData = append(csvData, vars[11].String())                                              // Creation Timestamp
			csvData = append(csvData, vars[12].String())                                              // Resource Version
			csvData = append(csvData, vars[13].String())                                              // UID

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
