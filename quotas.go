package main

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

func convert2Milicores(value string) int {
	if strings.Contains(value, "m") {
		intValue, _ := strconv.Atoi(strings.TrimSuffix(value, "m"))
		return intValue
	}
	intValue, _ := strconv.Atoi(value)
	return intValue * 1000
}

// TODO: Support for Tb, Gb, Mb as well
func convert2Gibibytes(value string) float64 {
	if i, err := strconv.ParseFloat(value, 64); err == nil {
		if i == 0 {
			return 0
		}
		return math.Round(((i / math.Pow(1024, 3)) * 100) / 100)
	}
	if strings.Contains(value, "Gi") {
		intValue, _ := strconv.ParseFloat(strings.TrimSuffix(value, "Gi"), 64)
		return intValue
	}
	if strings.Contains(value, "Ti") {
		intValue, _ := strconv.ParseFloat(strings.TrimSuffix(value, "Ti"), 64)
		return intValue * math.Pow(1024, 1)
	}
	if strings.Contains(value, "Mi") {
		intValue, _ := strconv.ParseFloat(strings.TrimSuffix(value, "Mi"), 64)
		return math.Round(((intValue / math.Pow(1024, 1)) * 100) / 100)
	}
	if strings.Contains(value, "Ki") {
		intValue, _ := strconv.ParseFloat(strings.TrimSuffix(value, "Ki"), 64)
		return math.Round(((intValue / math.Pow(1024, 2)) * 100) / 100)
	}
	return -1
}

func getNamespaceQuotas() {

	var csvHeader = []string{"Cluster", "Namespace", "Hard.CPUReq", "Hard.CPULim", "Hard.MEMReq", "Hard.MEMLim",
		"Used.CPUReq", "Used.CPULim", "Used.MEMReq", "Used.MEMLim", "CreationTime", "Version", "UID"}
	var csvData []interface{}
	var startTime time.Time
	var duration time.Duration
	var xR int = 1
	//var cell string = "A1"
	var sheetName string = "NSQuotas"
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

		apiurl = cfg.Clusters[i].BaseURL + "/api/v1/resourcequotas?limit=1000"
		body, _ := getRest(apiurl, cfg.Clusters[i].Token)

		// Loop in list items and export data into Excel file
		items := gjson.GetBytes(body, "items")
		items.ForEach(func(key, value gjson.Result) bool {
			vars := gjson.GetMany(value.String(),
				`metadata.name`, `metadata.namespace`, `spec.hard.requests\.cpu`, `spec.hard.limits\.cpu`, `spec.hard.requests\.memory`, `spec.hard.limits\.memory`,
				`status.used.requests\.cpu`, `status.used.limits\.cpu`, `status.used.requests\.memory`, `status.used.limits\.memory`,
				`metadata.creationTimestamp`, `metadata.resourceVersion`, `metadata.uid`)

			if vars[0].String() != cfg.Clusters[i].Quota {
				return true
			}

			csvData = nil
			csvData = append(csvData, cfg.Clusters[i].Name)                // Cluster Name
			csvData = append(csvData, vars[1].String())                    // Namespace Name
			csvData = append(csvData, convert2Milicores(vars[2].String())) // Hard CPU Requests
			csvData = append(csvData, convert2Milicores(vars[3].String())) // Hard CPU Limits
			csvData = append(csvData, convert2Gibibytes(vars[4].String())) // Hard MEM Requests
			csvData = append(csvData, convert2Gibibytes(vars[5].String())) // Hard MEM Limits
			csvData = append(csvData, convert2Milicores(vars[6].String())) // Used CPU Requests
			csvData = append(csvData, convert2Milicores(vars[7].String())) // Used CPU Limits
			csvData = append(csvData, convert2Gibibytes(vars[8].String())) // Used MEM Requests
			csvData = append(csvData, convert2Gibibytes(vars[9].String())) // Used MEM Limits
			csvData = append(csvData, vars[10].String())                   // Creation Timestamp
			csvData = append(csvData, vars[11].String())                   // Resource Version
			csvData = append(csvData, vars[12].String())                   // UID

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
