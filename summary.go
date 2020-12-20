package main

func createSummarySheet() {
	var sheet string = "Summary"
	//var datasheet string
	//var col, row string

	// Initialize Excel sheet
	xf.SetActiveSheet(xf.GetSheetIndex(sheet))

	xf.SetCellValue(sheet, "B2", "This summary sheet has been left blank intentionally")
	/*
		datasheet = "Alerts"
		err := xf.AddPivotTable(&excelize.PivotTableOption{
			DataRange:           datasheet + "!$A$1:" + getRightBottomCell(datasheet),
			PivotTableRange:     sheet + "!$B$1:$E$1",
			Rows:                []excelize.PivotTableField{{Data: "Cluster", DefaultSubtotal: true}},
			Columns:             []excelize.PivotTableField{{Data: "Severity", DefaultSubtotal: true}},
			Data:                []excelize.PivotTableField{{Data: "Severity", Name: "Severity", Subtotal: "Count"}},
			RowGrandTotals:      false,
			ColGrandTotals:      true,
			ShowDrill:           true,
			ShowRowHeaders:      true,
			ShowColHeaders:      true,
			ShowLastColumn:      true,
			PivotTableStyleName: "PivotStyleMedium14",
		})
		perror(err)

		datasheet = "Namespaces"
		err = xf.AddPivotTable(&excelize.PivotTableOption{
			DataRange:       datasheet + "!$A$1:" + getRightBottomCell(datasheet),
			PivotTableRange: sheet + "!$B$10:$E$10",
			Rows:            []excelize.PivotTableField{{Data: "Cluster", DefaultSubtotal: true}},
			Data: []excelize.PivotTableField{
				{Data: "Name", Name: "Namespaces", Subtotal: "Count"},
				{Data: "TotalPods", Name: "TotPods", Subtotal: "Sum"},
				{Data: "RunningPods", Name: "RunPods", Subtotal: "Sum"}},
			RowGrandTotals:      false,
			ColGrandTotals:      true,
			ShowDrill:           true,
			ShowRowHeaders:      true,
			ShowColHeaders:      true,
			ShowLastColumn:      true,
			PivotTableStyleName: "PivotStyleMedium14",
		})
		perror(err)  */
}
