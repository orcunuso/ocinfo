package main

import "strconv"

func printCounts(sheet, cell, description, targetSheet string) {
	var rowCount string
	rows, _ := xf.GetRows(targetSheet)
	if len(rows) == 0 {
		rowCount = "N/A"
	} else {
		rowCount = strconv.Itoa(len(rows) - 1)
	}
	xf.SetSheetRow(sheet, cell, &[]string{description, rowCount})
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
		"fill": {"type": "pattern","color": ["#FFCCCC"],"pattern": 1}}`)

	xsr, _ := xf.NewStyle(`{
		"font": {"family": "Consolas", "size": 12, "color": "#000000", "bold": false},
		"alignment": {"horizontal": "center", "vertical": "center"},
		"border": [
			{"type": "left", "color": "000000", "style": 1}, 
			{"type": "right", "color": "000000", "style": 1}, 
			{"type": "top", "color": "000000", "style": 1}, 
			{"type": "bottom", "color": "000000", "style": 1}],
		"fill": {"type": "pattern","color": ["#FFFFFF"],"pattern": 1}}`)

	xf.SetColWidth(sheet, "A", "A", 2)
	printCounts(sheet, "B2", "Total Count of Clusters", "Clusters")
	printCounts(sheet, "B3", "Total Count of Nodes", "Nodes")
	printCounts(sheet, "B4", "Total Count of Namespaces", "Namespaces")
	printCounts(sheet, "B5", "Total Count of Persistent Volumes", "PVolumes")
	printCounts(sheet, "B6", "Total Count of DaemonSets", "DaemonSets")
	printCounts(sheet, "B7", "Total Count of Routes", "Routes")
	printCounts(sheet, "B8", "Total Count of Services", "Services")
	printCounts(sheet, "B9", "Total Count of Deployments", "Deployments")
	printCounts(sheet, "B10", "Total Count of Running Pods", "Pods")
	xf.SetCellValue(sheet, "B12", "Creation Timestamp: "+currentTime.Format("2006-01-02T15:04:05"))
	xf.SetCellStyle(sheet, "B2", "B10", xsl)
	xf.SetCellStyle(sheet, "C2", "C10", xsr)
	xf.SetCellStyle(sheet, "B12", "B12", xsr2)
	autoFit(xf, "Summary", 1.25, 16)
}
