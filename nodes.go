package main

import (
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

func getNodeRole(labels string) string {
	if strings.Contains(labels, "node-role.kubernetes.io/master") {
		return "master"
	}
	return "worker"
}

func getNodeStatus(status string) string {
	if status == "True" {
		return "Ready"
	}
	return "NotReady"
}

func getNodes() {

	var csvHeader = []string{"Cluster", "Name", "Role", "Status", "IP", "PodCIDR", "CPU", "MEM", "KubeletVer", "MachineName", "CreationTime", "Version", "UID"}
	var csvData []string
	var startTime time.Time
	var duration time.Duration
	var xR int = 1
	var sheetName string = "Nodes"
	var apiurlNod string

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

		apiurlNod = cfg.Clusters[i].BaseURL + "/api/v1/nodes?limit=1000"
		rBody, _ := getRest(apiurlNod, cfg.Clusters[i].Token)

		// Loop in list items and export data into Excel file
		items := gjson.GetBytes(rBody, "items")
		items.ForEach(func(key, value gjson.Result) bool {
			vars := gjson.GetMany(value.String(),
				`metadata.name`, `metadata.labels`, `status.conditions.#(type=="Ready").status`, `status.addresses.#(type=="InternalIP").address`, `spec.podCIDR`,
				`status.allocatable.cpu`, `status.allocatable.memory`, `status.nodeInfo.kubeletVersion`, `metadata.annotations.machine\.openshift\.io/machine`,
				`metadata.creationTimestamp`, `metadata.resourceVersion`, `metadata.uid`)

			csvData = nil
			csvData = append(csvData, cfg.Clusters[i].Name)                    // Cluster Name
			csvData = append(csvData, vars[0].String())                        // Node Name
			csvData = append(csvData, getNodeRole(vars[1].String()))           // Node Role
			csvData = append(csvData, getNodeStatus(vars[2].String()))         // Status
			csvData = append(csvData, vars[3].String())                        // Node IP
			csvData = append(csvData, vars[4].String())                        // pod CIDR
			csvData = append(csvData, vars[5].String())                        // Node CPU
			csvData = append(csvData, vars[6].String())                        // Node MEM
			csvData = append(csvData, vars[7].String())                        // Kubelet Version
			csvData = append(csvData, strings.Split(vars[8].String(), "/")[1]) // Machine Name
			csvData = append(csvData, vars[9].String())                        // Creation Timestamp
			csvData = append(csvData, vars[10].String())                       // Resource Version
			csvData = append(csvData, vars[11].String())                       // UID

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
