package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

var (
	warn *log.Logger
	info *log.Logger
	erro *log.Logger
)

var (
	black   = color("\033[1;30m%s\033[0m")
	red     = color("\033[1;31m%s\033[0m")
	green   = color("\033[1;32m%s\033[0m")
	yellow  = color("\033[1;33m%s\033[0m")
	purple  = color("\033[1;34m%s\033[0m")
	magenta = color("\033[1;35m%s\033[0m")
	teal    = color("\033[1;36m%s\033[0m")
	white   = color("\033[1;37m%s\033[0m")
)

func color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString, fmt.Sprint(args...))
	}
	return sprint
}

func perror(err error) {
	if err != nil {
		erro.Println(err.Error())
	}
}

// This function returns response body as a slice of bytes and status code of a GET request.
func getRest(url string, token string) ([]byte, int) {
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("Authorization", "Bearer "+token)

	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: transport}
	response, err := client.Do(request)
	perror(err)
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	perror(err)
	return body, response.StatusCode
}

func formatDate(d string) string {
	t, err := time.Parse(time.RFC3339, d)
	if err != nil {
		erro.Println(err)
	}
	return t.Format("2006.01.02")
}

// This function checks if OpenShift API is accessable.
func checkClusterAPI(apiurl string, token string) (bool, string) {
	request, err := http.NewRequest("GET", apiurl+"/api/v1", nil)
	request.Header.Add("Authorization", "Bearer "+token)

	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: transport}
	response, err := client.Do(request)

	if err != nil {
		return false, err.Error()
	}

	if response.StatusCode >= 400 {
		return false, http.StatusText(response.StatusCode)
	}
	return true, "Success"
}

// This function "tries" to autofit the width of all columns in a sheet. Normally this is supposed to be a more complicated algorithm
// as the width of all characters in all font types greatly differ so this is the best it can get with minimal effort (especially with
// monospaced fonts like Consolas)
func autoFit(f *excelize.File, sheet string) {
	var max int = 0
	var ilk int = 0
	var i int = 1
	cols, _ := f.Cols(sheet)

	for cols.Next() {
		max = 0
		rows, _ := cols.Rows()
		for x, c := range rows {
			if x == 0 {
				ilk = len(c)
			}
			if len(c) > max {
				max = len(c)
				if max > 66 {
					max = 66
					break
				}
			}
		}
		cn, _ := excelize.ColumnNumberToName(i)

		if ilk == max {
			f.SetColWidth(sheet, cn, cn, float64(max)+3)
		} else {
			f.SetColWidth(sheet, cn, cn, float64(max)+1)
		}
		i++
	}
}

// This function retuns the address of the lower-end, corner cell within a range of cells.
func getRightBottomCell(sheet string) string {
	cl, _ := xf.GetCols(sheet)
	rw, _ := xf.GetRows(sheet)
	c, _ := excelize.CoordinatesToCellName(len(cl), len(rw))
	return c
}

// This function turns a range of cells into a nice, human-readable table format with filters, freeze panes and different row styles.
func formatTable(sheet string, colCount int) {
	var i int = 1
	var c1, c2 string
	index := xf.GetSheetIndex(sheet)
	xf.SetActiveSheet(index)

	rows, _ := xf.Rows(sheet)
	for rows.Next() {
		if i == 1 {
			c2, _ = excelize.CoordinatesToCellName(colCount, i)
			xf.SetCellStyle(sheet, "A1", c2, xsh0)
			xf.AutoFilter(sheet, "A1", c2, "")
		} else {
			if (i % 2) == 0 {
				c1, _ = excelize.CoordinatesToCellName(1, i)
				c2, _ = excelize.CoordinatesToCellName(colCount, i)
				xf.SetCellStyle(sheet, c1, c2, xsr2)
			} else {
				c1, _ = excelize.CoordinatesToCellName(1, i)
				c2, _ = excelize.CoordinatesToCellName(colCount, i)
				xf.SetCellStyle(sheet, c1, c2, xsr1)
			}
		}
		i++
	}

	xf.SetPanes(sheet, `{
		"freeze": true, "split": false, "x_split": 2, "y_split": 1,
		"top_left_cell": "C2", "active_pane": "bottomRight",
		"panes": [
		{
			"sqref": "C2",
			"active_cell": "C2",
			"pane": "bottomLeft"
		}]
	}`)

	autoFit(xf, sheet)
}

// Conditional formatting function that takes the column and a slice of strings as conditions to be formatted. The values of the slice
// need to be in order of the color formats defined as red, yellow, green, blue and gray.
// Ex: conditionalFormat5("Sheet1", "E", []string{"critical", "warning", "info", "trivial", "none"})
func conditionalFormat5(sheet, col string, values []string) {
	var formats []int
	format1, _ := xf.NewConditionalStyle(`{"font":{"color": "#9A0511", "size": 10},"fill":{"type": "pattern","color": ["#FEC7CE"],"pattern": 1}}`) // Light red
	format2, _ := xf.NewConditionalStyle(`{"font":{"color": "#9B5713", "size": 10},"fill":{"type": "pattern","color": ["#FEEAA0"],"pattern": 1}}`) // Light yellow
	format3, _ := xf.NewConditionalStyle(`{"font":{"color": "#09600B", "size": 10},"fill":{"type": "pattern","color": ["#C7EECF"],"pattern": 1}}`) // Light green
	format4, _ := xf.NewConditionalStyle(`{"font":{"color": "#0033CC", "size": 10},"fill":{"type": "pattern","color": ["#99CCFF"],"pattern": 1}}`) // Light blue
	format5, _ := xf.NewConditionalStyle(`{"font":{"color": "#454545", "size": 10},"fill":{"type": "pattern","color": ["#E6E6E7"],"pattern": 1}}`) // Light gray
	formats = append(formats, format1, format2, format3, format4, format5)

	rows, _ := xf.GetRows(sheet)
	cells := col + "1:" + col + fmt.Sprint(len(rows))

	for i, value := range values {
		err := xf.SetConditionalFormat(sheet, cells, fmt.Sprintf(`[{
			"type": "cell", "criteria": "==", "format": %d, "value": "\"%s\""}]`, formats[i], value))
		if err != nil {
			fmt.Println(err)
		}
	}
}
