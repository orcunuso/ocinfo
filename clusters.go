package main

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

func getNodeCount(apiurl string, token string) int {
	nodeCount := 0
	body, status := getRest(apiurl, token)
	if status == 404 {
		return 0
	}

	items := gjson.GetBytes(body, "items")
	items.ForEach(func(key, value gjson.Result) bool {
		nodeCount++
		return true
	})
	return nodeCount
}

func getIngressIP(apiurl string) string {
	temp := strings.Split(apiurl, ":")[1]
	temp = strings.TrimPrefix(temp, "//")
	temp = strings.Replace(temp, "api", "console-openshift-console.apps", 1)
	ips, err := net.LookupHost(temp)
	if err != nil {
		return "NoHost"
	}
	return ips[0]
}

func getAvailableVersion(result gjson.Result) string {
	var sv string = "x.y.0"
	result.ForEach(func(key, value gjson.Result) bool {
		sx := gjson.Get(value.String(), `version`)
		ix, _ := strconv.Atoi(strings.Split(sx.String(), ".")[2])
		iv, _ := strconv.Atoi(strings.Split(sv, ".")[2])
		if ix > iv {
			sv = sx.String()
		}
		return true
	})
	return sv
}

func getClusters() {
	var csvHeader = []string{"Cluster", "APIURL", "Version", "Channel", "Available", "Nodes", "IngressIP", "CNI", "Pod CIDR", "ServiceCIDR", "HostPrefix", "ClusterID"}
	var csvData []interface{}
	var startTime time.Time
	var duration time.Duration
	var xR int = 1
	var sheetName string = "Clusters"
	var apiurlVer, apiurlNod, apiurlNet string

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

		apiurlVer = cfg.Clusters[i].BaseURL + "/apis/config.openshift.io/v1/clusterversions?limit=1000"
		apiurlNod = cfg.Clusters[i].BaseURL + "/api/v1/nodes?limit=1000"
		apiurlNet = cfg.Clusters[i].BaseURL + "/api/v1/namespaces/openshift-network-operator/configmaps/applied-cluster"
		rBody, _ := getRest(apiurlVer, cfg.Clusters[i].Token)
		nBody, _ := getRest(apiurlNet, cfg.Clusters[i].Token)

		// Loop in list items and export data into Excel file
		items := gjson.GetBytes(rBody, "items")
		items.ForEach(func(key, value gjson.Result) bool {
			vars := gjson.GetMany(value.String(), `spec.desiredUpdate.version`, `spec.channel`, `status.availableUpdates`, `spec.clusterID`)
			data := gjson.GetBytes(nBody, "data.applied")

			csvData = nil
			csvData = append(csvData, cfg.Clusters[i].Name)                                    // Cluster Name
			csvData = append(csvData, cfg.Clusters[i].BaseURL)                                 // API URL
			csvData = append(csvData, vars[0].String())                                        // Cluster Version
			csvData = append(csvData, vars[1].String())                                        // Cluster Update Channel
			csvData = append(csvData, getAvailableVersion(vars[2]))                            // Available Updates
			csvData = append(csvData, getNodeCount(apiurlNod, cfg.Clusters[i].Token))          // Number of Nodes
			csvData = append(csvData, getIngressIP(cfg.Clusters[i].BaseURL))                   // Default Ingress IP
			csvData = append(csvData, gjson.Get(data.String(), "defaultNetwork.type"))         // CNI Type
			csvData = append(csvData, gjson.Get(data.String(), "clusterNetwork.0.cidr"))       // Pod CIDR
			csvData = append(csvData, gjson.Get(data.String(), "serviceNetwork.0"))            // Service CIDR
			csvData = append(csvData, gjson.Get(data.String(), "clusterNetwork.0.hostPrefix")) // Host Prefix
			csvData = append(csvData, vars[3].String())                                        // ClusterID

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
