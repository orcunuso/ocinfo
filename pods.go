package main

import (
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

func getPods() {
	var csvHeader = []string{"Cluster", "Namespace", "PodName", "ContainerName", "Phase", "PodIP", "NodeName", "#Restart", "#Sidecar", "#InitCon",
		"SCC", "ServiceAccount", "DNSPolicy", "RestartPolicy", "ImagePullPolicy", "Image", "RunAsUser", "qosClass",
		"CPU.Req", "CPU.Lim", "MEM.Req", "MEM.Lim", "CreationDate", "Version", "UID"}
	var csvData []interface{}
	var startTime time.Time
	var duration time.Duration
	var xR int = 1
	var sheetName string = "Pods"
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

		apiurl = cfg.Clusters[i].BaseURL + "/api/v1/pods"
		body, _ := getRest(apiurl, cfg.Clusters[i].Token)

		// Loop in list items and export data into Excel file
		items := gjson.GetBytes(body, "items")
		items.ForEach(func(key, value gjson.Result) bool {
			vars := gjson.GetMany(value.String(),
				`metadata.namespace`, `metadata.name`, `spec.containers.0.name`, `status.phase`, `status.podIP`, `spec.nodeName`,
				`status.containerStatuses.0.restartCount`, `spec.containers.#.name`, `spec.initContainers.#.name`,
				`metadata.annotations.openshift\.io/scc`, `spec.serviceAccountName`, `spec.dnsPolicy`, `spec.restartPolicy`,
				`spec.containers.0.imagePullPolicy`, `spec.containers.0.image`, `spec.containers.0.securityContext.runAsUser`, `status.qosClass`,
				`spec.containers.0.resources.requests.cpu`, `spec.containers.0.resources.limits.cpu`, `spec.containers.0.resources.requests.memory`,
				`spec.containers.0.resources.limits.memory`, `metadata.creationTimestamp`, `metadata.resourceVersion`, `metadata.uid`)

			if vars[3].String() != "Running" {
				return true
			}
			restartCount, _ := strconv.Atoi(vars[6].String())

			csvData = nil
			csvData = append(csvData, cfg.Clusters[i].Name)                 // Cluster Name
			csvData = append(csvData, vars[0].String())                     // Namespace Name
			csvData = append(csvData, vars[1].String())                     // Pod Name
			csvData = append(csvData, vars[2].String())                     // Main Container Name
			csvData = append(csvData, vars[3].String())                     // Pod Phase
			csvData = append(csvData, vars[4].String())                     // Pod IP Address
			csvData = append(csvData, vars[5].String())                     // Pod Node Name
			csvData = append(csvData, restartCount)                         // Main Container Restart Count
			csvData = append(csvData, len(vars[7].Array())-1)               // Pod Sidecar Container Count
			csvData = append(csvData, len(vars[8].Array()))                 // Pod Init Container Count
			csvData = append(csvData, vars[9].String())                     // Pod SCC
			csvData = append(csvData, vars[10].String())                    // Pod Service Account Name
			csvData = append(csvData, vars[11].String())                    // Pod DNS Policy
			csvData = append(csvData, vars[12].String())                    // Pod Restart Policy
			csvData = append(csvData, vars[13].String())                    // Main Container Image Pull Policy
			csvData = append(csvData, vars[14].String())                    // Main Container Image
			csvData = append(csvData, vars[15].String())                    // Main Container User
			csvData = append(csvData, vars[16].String())                    // Pod QOS Class
			csvData = append(csvData, convert2Milicores(vars[17].String())) // Main Container CPU Request
			csvData = append(csvData, convert2Milicores(vars[18].String())) // Main Container CPU Limit
			csvData = append(csvData, convert2Mebibytes(vars[19].String())) // Main Container MEM Request
			csvData = append(csvData, convert2Mebibytes(vars[20].String())) // Main Container MEM Limit
			csvData = append(csvData, formatDate(vars[21].String()))        // Creation Timestamp
			csvData = append(csvData, vars[22].String())                    // Resource Version
			csvData = append(csvData, vars[23].String())                    // UID

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
