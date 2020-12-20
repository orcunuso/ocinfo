package main

import (
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

func calculateCPUCores(core, socket string) int {
	x, _ := strconv.Atoi(core)
	y, _ := strconv.Atoi(socket)
	return x * y
}

func calculateMemoryGB(memory string) int {
	x, _ := strconv.Atoi(memory)
	return x / 1024
}

func createMachineSheet() {

	var csvHeader = []string{"Cluster", "Name", "Phase", "InstanceState", "NodeRef", "VmID", "ProviderID", "CPUCores", "RAMGB", "OSDisk", "CreationDate", "Version", "UID"}
	var csvData []interface{}
	var startTime time.Time
	var duration time.Duration
	var xR int = 1
	var sheetName string = "Machines"
	var apiurlMac string

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

		apiurlMac = cfg.Clusters[i].BaseURL + "/apis/machine.openshift.io/v1beta1/machines?limit=1000"
		rBody, _ := getRest(apiurlMac, cfg.Clusters[i].Token)

		// Loop in list items and export data into Excel file
		items := gjson.GetBytes(rBody, "items")
		items.ForEach(func(key, value gjson.Result) bool {
			vars := gjson.GetMany(value.String(),
				`metadata.name`, `status.phase`, `status.providerStatus.instanceState`, `status.nodeRef.name`, `metadata.annotations.VmId`, `spec.providerID`,
				`spec.providerSpec.value.cpu.cores`, `spec.providerSpec.value.cpu.sockets`, `spec.providerSpec.value.memory_mb`, `spec.providerSpec.value.os_disk.size_gb`,
				`metadata.creationTimestamp`, `metadata.resourceVersion`, `metadata.uid`)

			csvData = nil
			csvData = append(csvData, cfg.Clusters[i].Name)                                  // Cluster Name
			csvData = append(csvData, vars[0].String())                                      // Machine Name
			csvData = append(csvData, vars[1].String())                                      // Machine Phase
			csvData = append(csvData, vars[2].String())                                      // Instance State
			csvData = append(csvData, vars[3].String())                                      // Node Reference
			csvData = append(csvData, vars[4].String())                                      // VM ID
			csvData = append(csvData, vars[5].String())                                      // Provider ID
			csvData = append(csvData, calculateCPUCores(vars[6].String(), vars[7].String())) // CPU Cores
			csvData = append(csvData, calculateMemoryGB(vars[8].String()))                   // Memory
			csvData = append(csvData, vars[9].String())                                      // OS Disk Size
			csvData = append(csvData, formatDate(vars[10].String()))                         // Creation Timestamp
			csvData = append(csvData, vars[11].String())                                     // Resource Version
			csvData = append(csvData, vars[12].String())                                     // UID

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
