package main

import (
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

func getAccessModes(result gjson.Result) string {
	var modes string
	result.ForEach(func(key, value gjson.Result) bool {
		modes += value.String() + " "
		return true
	})
	modes = strings.Replace(modes, "ReadWriteMany", "RWX", 1)
	modes = strings.Replace(modes, "ReadOnlyMany", "ROX", 1)
	modes = strings.Replace(modes, "ReadWriteOnce", "RWO", 1)
	modes = strings.TrimSuffix(modes, " ")
	return modes
}

func createPVolumeSheet() {
	var csvHeader = []string{"Cluster", "Name", "Status", "Capacity", "AccessModes", "ReclaimPolicy", "ClaimedBy", "CreationDate", "Version", "UID"}
	var csvData []interface{}
	var startTime time.Time
	var duration time.Duration
	var xR int = 1
	var sheetName string = "PVolumes"
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

		apiurl = cfg.Clusters[i].BaseURL + "/api/v1/persistentvolumes?limit=1000"
		body, _ := getRest(apiurl, cfg.Clusters[i].Token)

		// Loop in list items and export data into Excel file
		items := gjson.GetBytes(body, "items")
		items.ForEach(func(key, value gjson.Result) bool {
			vars := gjson.GetMany(value.String(),
				`metadata.name`, `status.phase`, `spec.capacity.storage`, `spec.accessModes`, `spec.persistentVolumeReclaimPolicy`,
				`spec.claimRef.namespace`, `spec.claimRef.name`, `metadata.creationTimestamp`, `metadata.resourceVersion`, `metadata.uid`)

			csvData = nil
			csvData = append(csvData, cfg.Clusters[i].Name)                  // Cluster Name
			csvData = append(csvData, vars[0].String())                      // PVolume Name
			csvData = append(csvData, vars[1].String())                      // Status
			csvData = append(csvData, vars[2].String())                      // Disk Size
			csvData = append(csvData, getAccessModes(vars[3]))               // Access Modes
			csvData = append(csvData, vars[4].String())                      // Reclaim Policy
			csvData = append(csvData, vars[5].String()+"/"+vars[6].String()) // Claimed By
			csvData = append(csvData, formatDate(vars[7].String()))          // Creation Timestamp
			csvData = append(csvData, vars[8].String())                      // Resource Version
			csvData = append(csvData, vars[9].String())                      // UID

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
