package main

import (
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

type clusterMetric struct {
	prometrics map[string]prometric
}

type prometric struct {
	valuesCPU []float64
	valuesMEM []float64
}

func convert2Milicores(value string) int {
	if strings.Contains(value, "m") {
		intValue, _ := strconv.Atoi(strings.TrimSuffix(value, "m"))
		return intValue
	}
	intValue, _ := strconv.Atoi(value)
	return intValue * 1000
}

func convert2Mebibytes(value string) float64 {
	if value == "" {
		return 0
	}
	if i, err := strconv.ParseFloat(value, 64); err == nil { // Check if value includes any runes
		return math.Round(((i / math.Pow(1049, 2)) * 100) / 100)
	}

	if strings.Contains(value, "Gi") {
		intValue, _ := strconv.ParseFloat(strings.TrimSuffix(value, "Gi"), 64)
		return intValue * math.Pow(1024, 1)
	}
	if strings.Contains(value, "Ti") {
		intValue, _ := strconv.ParseFloat(strings.TrimSuffix(value, "Ti"), 64)
		return intValue * math.Pow(1024, 2)
	}
	if strings.Contains(value, "Mi") {
		intValue, _ := strconv.ParseFloat(strings.TrimSuffix(value, "Mi"), 64)
		return intValue
	}
	if strings.Contains(value, "Ki") {
		intValue, _ := strconv.ParseFloat(strings.TrimSuffix(value, "Ki"), 64)
		return math.Round(((intValue / math.Pow(1024, 1)) * 100) / 100)
	}
	if strings.Contains(value, "T") {
		intValue, _ := strconv.ParseFloat(strings.TrimSuffix(value, "T"), 64)
		return math.Round(intValue * math.Pow(1000, 2) / 1.049)
	}
	if strings.Contains(value, "G") {
		intValue, _ := strconv.ParseFloat(strings.TrimSuffix(value, "G"), 64)
		return math.Round(intValue * math.Pow(1000, 1) / 1.049)
	}
	if strings.Contains(value, "M") {
		intValue, _ := strconv.ParseFloat(strings.TrimSuffix(value, "M"), 64)
		return math.Round(intValue / 1.049)
	}
	if strings.Contains(value, "K") {
		intValue, _ := strconv.ParseFloat(strings.TrimSuffix(value, "K"), 64)
		return math.Round(intValue / 1049)
	}
	return -1
}

// This function queries default Prometheus instance and returns utilization of namespaces in different time ranges (5m, 1h, 8h)
func (cm *clusterMetric) queryPrometheus(baseurl, token string) {
	var trs = []string{"5m", "1h", "8h"}
	var tri = 0
	promurl := "https://prometheus-k8s-openshift-monitoring.apps." + strings.TrimLeft(strings.Split(baseurl, ":")[1], "//api.")
	promurl += "/api/v1/query?query=x"

	for _, tr := range trs {
		// Query CPU usage
		u, err := url.Parse(promurl)
		perror(err, false)
		q := u.Query()
		q.Set("query", "sum(rate(container_cpu_usage_seconds_total{image!=\"\"}["+tr+"])) by (namespace)")
		u.RawQuery = q.Encode()

		body, code := getRest(u.String(), token)
		if code >= 400 {
			erro.Printf("Prometheus query failed: %s", http.StatusText(code))
			break
		} else {
			if (gjson.GetBytes(body, "status")).String() != "success" {
				erro.Printf("Prometheus query failed for unknown reason")
				break
			}
		}

		items := gjson.GetBytes(body, "data.result")
		items.ForEach(func(key, value gjson.Result) bool {
			ns := gjson.Get(value.String(), `metric.namespace`)
			va := gjson.Get(value.String(), `value.1`)
			_, found := cm.prometrics[ns.String()]
			if found {
				cm.prometrics[ns.String()].valuesCPU[tri] = math.Round(va.Float() * 1000)
			} else {
				scpu := make([]float64, 3)
				smem := make([]float64, 3)
				pm := prometric{scpu, smem}
				pm.valuesCPU[tri] = math.Round(va.Float() * 1000)
				cm.prometrics[ns.String()] = pm
			}
			return true
		})
		// Query Memory usage
		q.Set("query", "max_over_time(sum(container_memory_working_set_bytes{image!=\"\"} / 1024^2) by (namespace)["+tr+":])")
		u.RawQuery = q.Encode()

		body, code = getRest(u.String(), token)
		if code >= 400 {
			erro.Printf("Prometheus query failed: %s", http.StatusText(code))
			break
		} else {
			if (gjson.GetBytes(body, "status")).String() != "success" {
				erro.Printf("Prometheus query failed for unknown reason")
				break
			}
		}

		items = gjson.GetBytes(body, "data.result")
		items.ForEach(func(key, value gjson.Result) bool {
			ns := gjson.Get(value.String(), `metric.namespace`)
			va := gjson.Get(value.String(), `value.1`)
			_, found := cm.prometrics[ns.String()]
			if found {
				cm.prometrics[ns.String()].valuesMEM[tri] = math.Round(va.Float()*10) / 10
			} else {
				scpu := make([]float64, 3)
				smem := make([]float64, 3)
				pm := prometric{scpu, smem}
				pm.valuesMEM[tri] = math.Round(va.Float()*10) / 10
				cm.prometrics[ns.String()] = pm
			}
			return true
		})
		tri++
	}
}

func createQuotaSheet() {

	var csvHeader = []string{"Cluster", "Namespace", "Hard.CPUReq", "Hard.CPULim", "Hard.MEMReq", "Hard.MEMLim", "Used.CPUReq", "Used.CPULim",
		"Used.MEMReq", "Used.MEMLim", "Real.CPU.5m", "Real.CPU.1h", "Real.CPU.8h", "Real.MEM.5m", "Real.MEM.1h", "Real.MEM.8h",
		"CreationDate", "Version", "UID"}
	var csvData []interface{}
	var startTime time.Time
	var duration time.Duration
	var xR int = 1
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

		var cms clusterMetric
		if cfg.Clusters[i].PromToken == "" {
			warn.Printf("Prometheus token is empty. Alerts sheet needs to be enabled in %s to get the token", ff)
		} else {
			cms.prometrics = make(map[string]prometric)
			cms.queryPrometheus(cfg.Clusters[i].BaseURL, cfg.Clusters[i].PromToken)
			//for k, v := range cms.prometrics {
			//	fmt.Println(k, "value is", v)
			//}
		}

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
			csvData = append(csvData, convert2Mebibytes(vars[4].String())) // Hard MEM Requests
			csvData = append(csvData, convert2Mebibytes(vars[5].String())) // Hard MEM Limits
			csvData = append(csvData, convert2Milicores(vars[6].String())) // Used CPU Requests
			csvData = append(csvData, convert2Milicores(vars[7].String())) // Used CPU Limits
			csvData = append(csvData, convert2Mebibytes(vars[8].String())) // Used MEM Requests
			csvData = append(csvData, convert2Mebibytes(vars[9].String())) // Used MEM Limits
			if v, found := cms.prometrics[vars[1].String()]; found {       // Real CPU and MEM Usage (5m, 1h, 8h)
				for _, fval := range v.valuesCPU {
					csvData = append(csvData, fval)
				}
				for _, fval := range v.valuesMEM {
					csvData = append(csvData, fval)
				}
			} else {
				csvData = append(csvData, 0, 0, 0, 0, 0, 0)
			}
			csvData = append(csvData, formatDate(vars[10].String())) // Creation Timestamp
			csvData = append(csvData, vars[11].String())             // Resource Version
			csvData = append(csvData, vars[12].String())             // UID

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
