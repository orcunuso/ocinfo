package main

import (
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

func getEgressIP(body []byte, namespace string) string {
	var returnValue string = "nil"
	items := gjson.GetBytes(body, "items")
	items.ForEach(func(key, value gjson.Result) bool {
		vars := gjson.GetMany(value.String(), `metadata.name`, `egressIPs.0`)
		if vars[0].String() == namespace {
			returnValue = vars[1].String()
			return false
		}
		return true
	})
	return returnValue
}

func getNsType(displayName string) string {
	if strings.HasPrefix(displayName, "TURKCELL") {
		return "application"
	}
	return "system"
}

func getPodCount(apiurl string, token string) (int, int) {
	podTotal, podRunning := 0, 0
	body, _ := getRest(apiurl, token)
	items := gjson.GetBytes(body, "items")
	items.ForEach(func(key, value gjson.Result) bool {
		podTotal++
		phase := gjson.Get(value.String(), "status.phase")
		if phase.String() == "Running" {
			podRunning++
		}
		return true
	})
	return podTotal, podRunning
}

func getNamespaces() {

	var csvHeader = []string{"Cluster", "Name", "Type", "DisplayName", "Description", "Requester", "NodeSelector", "SCC.MCS", "SCC.UIDRange", "RequestID", "ServiceID",
		"EgressIP", "TotalPods", "RunningPods", "CreationTime", "Version", "UID"}
	var csvData []interface{}
	var startTime time.Time
	var duration time.Duration
	var xR int = 1
	var sheetName string = "Namespaces"
	var apiurl, apiurlNet string

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

		apiurl = cfg.Clusters[i].BaseURL + "/api/v1/namespaces"
		apiurlNet = cfg.Clusters[i].BaseURL + "/apis/network.openshift.io/v1/netnamespaces"
		body, _ := getRest(apiurl, cfg.Clusters[i].Token)
		bodyNet, statusNet := getRest(apiurlNet, cfg.Clusters[i].Token)

		// Loop in list items and export data into Excel file
		items := gjson.GetBytes(body, "items")
		items.ForEach(func(key, value gjson.Result) bool {
			vars := gjson.GetMany(value.String(),
				`metadata.name`, `metadata.annotations.openshift\.io/display-name`, `metadata.annotations.openshift\.io/description`,
				`metadata.annotations.openshift\.io/requester`, `metadata.annotations.openshift\.io/node-selector`,
				`metadata.annotations.openshift\.io/sa\.scc\.mcs`, `metadata.annotations.openshift\.io/sa\.scc\.uid-range`,
				`metadata.labels.tcrequestid`, `metadata.labels.tcserviceid`, `metadata.labels.`+cfg.Appnslabel,
				`metadata.creationTimestamp`, `metadata.resourceVersion`, `metadata.uid`)

			ns := vars[0].String()
			pt, pr := getPodCount(apiurl+"/"+ns+"/pods", cfg.Clusters[i].Token)

			csvData = nil
			csvData = append(csvData, cfg.Clusters[i].Name) // Cluster Name
			csvData = append(csvData, ns)                   // Namespace Name
			if vars[9].String() != "" {                     // Namespace Type
				csvData = append(csvData, "application")
			} else {
				csvData = append(csvData, "system")
			}
			csvData = append(csvData, vars[1].String()) // Namespace Display Name
			csvData = append(csvData, vars[2].String()) // Namespace Description
			csvData = append(csvData, vars[3].String()) // Namespace Requester
			csvData = append(csvData, vars[4].String()) // Namespace NodeSelector
			csvData = append(csvData, vars[5].String()) // Namespace SCC.MCS
			csvData = append(csvData, vars[6].String()) // Namespace SCC.UID-Range
			csvData = append(csvData, vars[7].String()) // Company Specific Label: RequestID
			csvData = append(csvData, vars[8].String()) // Company Specific Label: ServiceID
			if statusNet == 404 {                       // This probably means that OpenShiftSDN is not implemented
				csvData = append(csvData, "NotImplemented")
			} else {
				csvData = append(csvData, getEgressIP(bodyNet, ns))
			}
			csvData = append(csvData, pt)                // Total Pod Count
			csvData = append(csvData, pr)                // Running Pod Count
			csvData = append(csvData, vars[10].String()) // Creation Timestamp
			csvData = append(csvData, vars[11].String()) // Resource Version
			csvData = append(csvData, vars[12].String()) // UID

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
