package main

import (
	"strconv"
)

func printTotalCounts(sheet, cell, description, targetSheet string) {
	var rowCount string
	rows, _ := xf.GetRows(targetSheet)
	if len(rows) == 0 {
		rowCount = "N/A"
	} else {
		rowCount = strconv.Itoa(len(rows) - 1)
	}
	xf.SetSheetRow(sheet, cell, &[]string{description, rowCount})
}

func printCondCounts(sheet, cell, description, targetSheet, targetCol, condition string) {
	var count int = 0
	var xR int = 1
	rows, err := xf.Rows(targetSheet)
	if err != nil {
		xf.SetSheetRow(sheet, cell, &[]string{description, "N/A"})
		return
	}

	for rows.Next() {
		c := targetCol + strconv.Itoa(xR)
		value, err := xf.GetCellValue(targetSheet, c)
		if err != nil {
			erro.Println("GetCellValue:", err)
		} else {
			if value == condition {
				count++
			}
		}
		xR++
	}
	xf.SetSheetRow(sheet, cell, &[]string{description, strconv.Itoa(count)})
}

func createNodesChart(sheet, cell, description, targetSheet, targetCol string) {
	rows, _ := xf.GetRows(targetSheet)
	xR := strconv.Itoa(len(rows))

	xf.AddChart(sheet, cell, `{
        "type": "col",
        "series": [{"name": "`+targetSheet+`!$`+targetCol+`$1", "categories": "`+targetSheet+`!$A$2:$A$`+xR+`", 
			"values": "`+targetSheet+`!$`+targetCol+`$2:$`+targetCol+`$`+xR+`"}],
        "format": {"x_scale": 1.0, "y_scale": 1.0, "x_offset": 15, "y_offset": 10,
            "print_obj": true, "lock_aspect_ratio": false, "locked": false},
        "legend": {"position": "bottom", "show_legend_key": false},
        "title": {"name": "`+description+`"},
        "plotarea": {"show_bubble_size": true, "show_cat_name": false, "show_leader_lines": false,
            "show_percent": false, "show_series_name": false, "show_val": true},
        "show_blanks_as": "zero"
    }`)
}

func createSummarySheet() {
	var sheet string = "Summary"
	xf.SetActiveSheet(xf.GetSheetIndex(sheet))

	xsl, _ := xf.NewStyle(`{
		"font": {"family": "Consolas", "size": 12, "color": "#000000", "bold": true},
		"alignment": {"horizontal": "left", "vertical": "center"},
		"border": [
			{"type": "left", "color": "000000", "style": 1}, 
			{"type": "right", "color": "000000", "style": 1}, 
			{"type": "top", "color": "000000", "style": 1}, 
			{"type": "bottom", "color": "000000", "style": 1}],
		"fill": {"type": "pattern","color": ["#ADD8E6"],"pattern": 1}}`)

	xsr, _ := xf.NewStyle(`{
		"font": {"family": "Consolas", "size": 12, "color": "#000000", "bold": true},
		"alignment": {"horizontal": "center", "vertical": "center"},
		"border": [
			{"type": "left", "color": "000000", "style": 1}, 
			{"type": "right", "color": "000000", "style": 1}, 
			{"type": "top", "color": "000000", "style": 1}, 
			{"type": "bottom", "color": "000000", "style": 1}],
		"fill": {"type": "pattern","color": ["#FFFFFF"],"pattern": 1}}`)

	xf.SetColWidth(sheet, "A", "A", 2)
	printTotalCounts(sheet, "B2", "Total Count of Clusters", "Clusters")
	printTotalCounts(sheet, "B3", "Total Count of Nodes", "Nodes")
	printTotalCounts(sheet, "B4", "Total Count of Namespaces", "Namespaces")
	printTotalCounts(sheet, "B5", "Total Count of Persistent Volumes", "PVolumes")
	printTotalCounts(sheet, "B6", "Total Count of DaemonSets", "DaemonSets")
	printTotalCounts(sheet, "B7", "Total Count of Routes", "Routes")
	printTotalCounts(sheet, "B8", "Total Count of Services", "Services")
	printTotalCounts(sheet, "B9", "Total Count of Deployments", "Deployments")
	printTotalCounts(sheet, "B10", "Total Count of Running Pods", "Pods")
	xf.SetCellStyle(sheet, "B2", "B10", xsl)
	xf.SetCellStyle(sheet, "C2", "C10", xsr)

	printCondCounts(sheet, "B12", "Total Count of Critical Alerts", "Alerts", "C", "critical")
	printCondCounts(sheet, "B13", "Total Count of NotReady Nodes", "Nodes", "D", "NotReady")
	printCondCounts(sheet, "B14", "Total Count of Provisioned Machines", "Machines", "C", "Provisioned")
	printCondCounts(sheet, "B15", "Total Count of Failed PVs", "PVolumes", "C", "Failed")
	printCondCounts(sheet, "B16", "Total Count of NodePort Services", "Services", "D", "NodePort")
	xf.SetCellStyle(sheet, "B12", "B16", xsl)
	xf.SetCellStyle(sheet, "C12", "C16", xsr)

	xf.SetCellValue(sheet, "B18", "Creation Timestamp: "+currentTime.Format("2006-01-02T15:04:05"))
	xf.SetCellStyle(sheet, "B18", "B18", xsr2)
	autoFit(xf, "Summary", 1.33, 16)

	createNodesChart(sheet, "E2", "Number of Nodes", "Clusters", "H")
}
