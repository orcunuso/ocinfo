package main

import (
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

func getHAState(r string) string {
	x, err := strconv.Atoi(r)
	if err != nil {
		return "Error"
	}
	switch x {
	case 0:
		return "N/A"
	case 1:
		return "False"
	default:
		return "True"
	}
}

func getDeployments(sheet, typ string, i, xR int) {
	var csvData []interface{}
	var apiurl string

	if typ == "deploy" {
		apiurl = cfg.Clusters[i].BaseURL + "/apis/apps/v1/deployments"
	} else {
		apiurl = cfg.Clusters[i].BaseURL + "/apis/apps.openshift.io/v1/deploymentconfigs"
	}

	body, _ := getRest(apiurl, cfg.Clusters[i].Token)
	items := gjson.GetBytes(body, "items")
	items.ForEach(func(key, value gjson.Result) bool {
		vars := gjson.GetMany(value.String(),
			`metadata.namespace`, `metadata.name`, `status.readyReplicas`, `spec.replicas`, `spec.template.spec.containers.0.name`,
			`spec.template.spec.containers.#.name`, `spec.template.spec.initContainers.#.name`, `spec.strategy.type`,
			`spec.template.spec.serviceAccountName`, `spec.template.spec.dnsPolicy`, `spec.template.spec.restartPolicy`,
			`spec.template.spec.containers.0.imagePullPolicy`, `spec.template.spec.containers.0.image`,
			`spec.template.spec.containers.0.livenessProbe`, `spec.template.spec.containers.0.readinessProbe`,
			`spec.template.spec.containers.0.resources.requests.cpu`, `spec.template.spec.containers.0.resources.limits.cpu`,
			`spec.template.spec.containers.0.resources.requests.memory`, `spec.template.spec.containers.0.resources.limits.memory`,
			`metadata.creationTimestamp`, `metadata.resourceVersion`, `metadata.uid`)

		readyReplicas := "0"
		if vars[2].String() != "" {
			readyReplicas = vars[2].String()
		}

		csvData = nil
		csvData = append(csvData, cfg.Clusters[i].Name)                 // Cluster Name
		csvData = append(csvData, vars[0].String())                     // Namespace Name
		csvData = append(csvData, vars[1].String())                     // Deployment Name
		csvData = append(csvData, typ)                                  // Deployment Type
		csvData = append(csvData, readyReplicas+"/"+vars[3].String())   // Ready Replicas
		csvData = append(csvData, getHAState(vars[3].String()))         // HA State
		csvData = append(csvData, vars[4].String())                     // Main Container Name
		csvData = append(csvData, len(vars[5].Array())-1)               // Pod Sidecar Container Count
		csvData = append(csvData, len(vars[6].Array()))                 // Pod Init Container Count
		csvData = append(csvData, vars[7].String())                     // Deployment Update Strategy
		csvData = append(csvData, vars[8].String())                     // Deployment Servica Account Name
		csvData = append(csvData, vars[9].String())                     // Pod DNS Policy
		csvData = append(csvData, vars[10].String())                    // Pod Restart Policy
		csvData = append(csvData, vars[11].String())                    // Main Container Image Pull Policy
		csvData = append(csvData, vars[12].String())                    // Main Container Image
		csvData = append(csvData, vars[13].String())                    // Main Container Liveness Probe
		csvData = append(csvData, vars[14].String())                    // Main Container Liveness Probe
		csvData = append(csvData, convert2Milicores(vars[15].String())) // Main Container CPU Request
		csvData = append(csvData, convert2Milicores(vars[16].String())) // Main Container CPU Limit
		csvData = append(csvData, convert2Mebibytes(vars[17].String())) // Main Container MEM Request
		csvData = append(csvData, convert2Mebibytes(vars[18].String())) // Main Container MEM Limit
		csvData = append(csvData, formatDate(vars[19].String()))        // Creation Timestamp
		csvData = append(csvData, vars[20].String())                    // Resource Version
		csvData = append(csvData, vars[21].String())                    // UID

		xR++
		cell, _ := excelize.CoordinatesToCellName(1, xR)
		xf.SetSheetRow(sheet, cell, &csvData)
		return true
	})
}

func createDeploymentSheet() {
	var csvHeader = []string{"Cluster", "Namespace", "Name", "Type", "Ready", "Concurrency", "ContainerName", "#Sidecar", "#InitCon", "UpdateStrategy",
		"ServiceAccount", "DNSPolicy", "RestartPolicy", "ImagePullPolicy", "Image", "LivenessProbe", "ReadinessProbe",
		"CPU.Req", "CPU.Lim", "MEM.Req", "MEM.Lim", "CreationDate", "Version", "UID"}
	var startTime time.Time
	var duration time.Duration
	var sheetName string = "Deployments"

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

		rows, _ := xf.GetRows(sheetName)
		getDeployments(sheetName, "deploy", i, len(rows))
		rows, _ = xf.GetRows(sheetName)
		getDeployments(sheetName, "dc", i, len(rows))
	}
	formatTable(sheetName, len(csvHeader))

	duration = time.Since(startTime)
	info.Printf("%s: Section ended in %s\n", sheetName, duration)
}
