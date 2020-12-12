// This sheet gets alerts from default Prometheus instance installed with OpenShift Monitoring Operator
package main

import (
	b64 "encoding/base64"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

func getAlerts() {
	var csvHeader = []string{"Cluster", "Alert", "Severity", "Instance", "StartTime", "EndTime", "Message", "State"}
	var csvData []string
	var startTime time.Time
	var duration time.Duration
	var xR int = 1
	var sheetName string = "Alerts"
	var apiurl string
	var secret, token string

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

		// Get secret name of prometheus-k8s service account
		apiurl = cfg.Clusters[i].BaseURL + "/api/v1/namespaces/openshift-monitoring/serviceaccounts/prometheus-k8s"
		body, _ := getRest(apiurl, cfg.Clusters[i].Token)
		items := gjson.GetBytes(body, "secrets")
		items.ForEach(func(key, value gjson.Result) bool {
			x := gjson.Get(value.String(), `name`)
			if strings.Contains(x.String(), "token") {
				secret = x.String()
				return false
			}
			return true
		})

		// Get token from secret resource
		apiurl = cfg.Clusters[i].BaseURL + "/api/v1/namespaces/openshift-monitoring/secrets/" + secret
		body, _ = getRest(apiurl, cfg.Clusters[i].Token)
		r := gjson.GetBytes(body, `data.token`)
		tb, _ := b64.StdEncoding.DecodeString(r.String())
		token = string(tb)

		// Get alerts from Prometheus with the token
		apiurl = "https://alertmanager-main-openshift-monitoring.apps." + strings.TrimLeft(strings.Split(cfg.Clusters[i].BaseURL, ":")[1], "//api.") + "/api/v1/alerts"
		body, _ = getRest(apiurl, token)

		// Loop in alerts data and export data into Excel file
		items = gjson.GetBytes(body, "data")
		items.ForEach(func(key, value gjson.Result) bool {
			vars := gjson.GetMany(value.String(), `labels.alertname`, `labels.severity`, `labels.instance`, `startsAt`, `endsAt`, `annotations.message`, `status.state`)

			csvData = nil
			csvData = append(csvData, cfg.Clusters[i].Name) // Cluster Name
			csvData = append(csvData, vars[0].String())     // Alert Name
			csvData = append(csvData, vars[1].String())     // Alert Severity
			csvData = append(csvData, vars[2].String())     // Alert Instance Name
			csvData = append(csvData, vars[3].String())     // Alert Start Time
			csvData = append(csvData, vars[4].String())     // Alert End Time
			csvData = append(csvData, vars[5].String())     // Alert Message
			csvData = append(csvData, vars[6].String())     // Alert State

			xR++
			cell, _ := excelize.CoordinatesToCellName(1, xR)
			xf.SetSheetRow(sheetName, cell, &csvData)
			return true
		})
	}
	formatTable(sheetName, len(csvHeader))
	conditionalFormat5(sheetName, "C", []string{"critical", "warning", "info", "undefined", "none"})

	duration = time.Since(startTime)
	info.Printf("%s: Section ended in %s\n", sheetName, duration)
}
