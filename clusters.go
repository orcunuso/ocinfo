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

func getInfraDetails(body []byte, i int) string {
	provider := gjson.GetBytes(body, `status.platformStatus.type`).String()
	var vars []gjson.Result

	// https://docs.openshift.com/container-platform/4.6/rest_api/config_apis/infrastructure-config-openshift-io-v1.html
	switch provider {
	case "oVirt":
		vars = gjson.GetManyBytes(body, `status.infrastructureName`, `status.platform`, `status.platformStatus.ovirt.apiServerInternalIP`,
			`status.platformStatus.ovirt.ingressIP`, `status.platformStatus.ovirt.nodeDNSIP`)
	case "VSphere":
		vars = gjson.GetManyBytes(body, `status.infrastructureName`, `status.platform`, `status.platformStatus.vsphere.apiServerInternalIP`,
			`status.platformStatus.vsphere.ingressIP`, `status.platformStatus.vsphere.nodeDNSIP`)
	case "OpenStack":
		vars = gjson.GetManyBytes(body, `status.infrastructureName`, `status.platform`, `status.platformStatus.openstack.apiServerInternalIP`,
			`status.platformStatus.openstack.ingressIP`, `status.platformStatus.openstack.nodeDNSIP`)
	case "BareMetal":
		vars = gjson.GetManyBytes(body, `status.infrastructureName`, `status.platform`, `status.platformStatus.baremetal.apiServerInternalIP`,
			`status.platformStatus.baremetal.ingressIP`, `status.platformStatus.baremetal.nodeDNSIP`)
	// For AWS, GCP and Azure, IP related fields will appear empty. I'll add provider specific specs in future releases.
	case "AWS":
		vars = gjson.GetManyBytes(body, `status.infrastructureName`, `status.platform`, `status.platformStatus.aws.apiServerInternalIP`,
			`status.platformStatus.aws.ingressIP`, `status.platformStatus.aws.nodeDNSIP`)
	case "GCP":
		vars = gjson.GetManyBytes(body, `status.infrastructureName`, `status.platform`, `status.platformStatus.gcp.apiServerInternalIP`,
			`status.platformStatus.gcp.ingressIP`, `status.platformStatus.gcp.nodeDNSIP`)
	case "Azure":
		vars = gjson.GetManyBytes(body, `status.infrastructureName`, `status.platform`, `status.platformStatus.azure.apiServerInternalIP`,
			`status.platformStatus.azure.ingressIP`, `status.platformStatus.azure.nodeDNSIP`)
	default:
		vars = gjson.GetManyBytes(body, `status.infrastructureName`, `status.platform`, `status.platformStatus.#.apiServerInternalIP`,
			`status.platformStatus.#.ingressIP`, `status.platformStatus.#.nodeDNSIP`)
	}
	return vars[i].String()
}

func createClusterSheet() {
	var csvHeader = []string{"Cluster", "APIServerURL", "Version", "InfraName", "Platform", "Channel", "Available", "#Nodes", "APIServerIP",
		"IngressIP", "NodeDNSIP", "CNI", "Pod CIDR", "ServiceCIDR", "HostPrefix", "ClusterID"}
	var csvData []interface{}
	var startTime time.Time
	var duration time.Duration
	var xR int = 1
	var sheetName string = "Clusters"
	var apiurlVer, apiurlNod, apiurlNet, apiurlInf string

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

		apiurlVer = cfg.Clusters[i].BaseURL + "/apis/config.openshift.io/v1/clusterversions?limit=1000"
		apiurlNod = cfg.Clusters[i].BaseURL + "/api/v1/nodes?limit=1000"
		apiurlNet = cfg.Clusters[i].BaseURL + "/api/v1/namespaces/openshift-network-operator/configmaps/applied-cluster"
		apiurlInf = cfg.Clusters[i].BaseURL + "/apis/config.openshift.io/v1/infrastructures/cluster"
		rBody, _ := getRest(apiurlVer, cfg.Clusters[i].Token)
		nBody, _ := getRest(apiurlNet, cfg.Clusters[i].Token)
		iBody, _ := getRest(apiurlInf, cfg.Clusters[i].Token)

		// Loop in list items and export data into Excel file
		items := gjson.GetBytes(rBody, "items")
		items.ForEach(func(key, value gjson.Result) bool {
			vars := gjson.GetMany(value.String(), `spec.desiredUpdate.version`, `spec.channel`, `status.availableUpdates`, `spec.clusterID`)
			data := gjson.GetBytes(nBody, "data.applied")

			cfg.Clusters[i].Provider = getInfraDetails(iBody, 1)
			if _, ok := testedOCPProviders[cfg.Clusters[i].Provider]; !ok {
				warn.Printf(yellow(cfg.Clusters[i].Name, "-> OCinfo is not well tested for this provider (", cfg.Clusters[i].Provider,
					"). You may expect empty values or inconsistent behaviors on provider specific fields"))
			}

			csvData = nil
			csvData = append(csvData, cfg.Clusters[i].Name)                                    // Cluster Name
			csvData = append(csvData, cfg.Clusters[i].BaseURL)                                 // API URL
			csvData = append(csvData, vars[0].String())                                        // Cluster Version
			csvData = append(csvData, getInfraDetails(iBody, 0))                               // Infrastructure Name
			csvData = append(csvData, cfg.Clusters[i].Provider)                                // Cloud Provider Name
			csvData = append(csvData, vars[1].String())                                        // Cluster Update Channel
			csvData = append(csvData, getAvailableVersion(vars[2]))                            // Available Updates
			csvData = append(csvData, getNodeCount(apiurlNod, cfg.Clusters[i].Token))          // Number of Nodes
			csvData = append(csvData, getInfraDetails(iBody, 2))                               // API Server IP
			csvData = append(csvData, getInfraDetails(iBody, 3))                               // Default Ingress IP
			csvData = append(csvData, getInfraDetails(iBody, 4))                               // Node DNS IP
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
