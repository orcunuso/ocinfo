package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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

func perror(err error, exit bool) {
	if err != nil {
		erro.Println(red(err.Error()))
		if exit {
			os.Exit(1)
		}
	}
}

// This function returns response body as a slice of bytes and status code of a GET request.
func getRest(url string, token string) ([]byte, int) {
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("Authorization", "Bearer "+token)

	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: transport}
	response, err := client.Do(request)
	perror(err, false)
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	perror(err, false)
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

// This function autofits the width of all columns in a sheet. Normally this is supposed to be a more complicated algorithm
// as the width of all characters in all font types greatly differ so this is the best it can get with minimal effort (especially with
// monospaced fonts like Consolas)
func autoFit(f *excelize.File, sheet string, width, height float64) {
	var max int = 0
	var ilk int = 0
	var i, j int = 1, 1
	cols, _ := f.Cols(sheet)
	rows, _ := f.Rows(sheet)

	for cols.Next() {
		max = 0
		rows, _ := cols.Rows()
		for x, c := range rows {
			if x == 0 {
				ilk = len(c)
			}
			if len(c) > max {
				max = len(c)
				if max > 80 {
					max = 80
					break
				}
			}
		}
		cn, _ := excelize.ColumnNumberToName(i)

		if ilk == max {
			f.SetColWidth(sheet, cn, cn, (float64(max)*width)+3)
		} else {
			f.SetColWidth(sheet, cn, cn, (float64(max)*width)+1)
		}
		i++
	}

	for rows.Next() {
		f.SetRowHeight(sheet, j, height)
		j++
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

	autoFit(xf, sheet, 0.9, 13)
}

// Conditional formatting function that takes the column and a slice of strings as conditions to be formatted. The values of the slice
// need to be in order of the color formats defined as red, yellow, green, blue and gray.
// Ex: conditionalFormat5("Sheet1", "E", []string{"critical", "warning", "info", "trivial", "none"})
func conditionalFormat5(sheet, col string, values []string) {
	var formats []int
	f1, _ := xf.NewConditionalStyle(`{"font":{"color": "#9A0511", "size": 10},"fill":{"type": "pattern","color": ["#FEC7CE"],"pattern": 1}}`) // Red
	f2, _ := xf.NewConditionalStyle(`{"font":{"color": "#9B5713", "size": 10},"fill":{"type": "pattern","color": ["#FEEAA0"],"pattern": 1}}`) // Yellow
	f3, _ := xf.NewConditionalStyle(`{"font":{"color": "#09600B", "size": 10},"fill":{"type": "pattern","color": ["#C7EECF"],"pattern": 1}}`) // Green
	f4, _ := xf.NewConditionalStyle(`{"font":{"color": "#0033CC", "size": 10},"fill":{"type": "pattern","color": ["#99CCFF"],"pattern": 1}}`) // Blue
	f5, _ := xf.NewConditionalStyle(`{"font":{"color": "#454545", "size": 10},"fill":{"type": "pattern","color": ["#E6E6E7"],"pattern": 1}}`) // Gray
	formats = append(formats, f1, f2, f3, f4, f5)

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

func createPivotTable(cell, sheet, col string) {
	var csvData []interface{}
	var csvHead = []string{""}
	var keys []string
	data := make(map[string]map[string]int)

	xC, _ := excelize.ColumnNameToNumber(col)
	rows, _ := xf.GetRows(sheet)
	for xR := 1; xR < len(rows); xR++ {
		c1, _ := excelize.CoordinatesToCellName(1, xR+1)
		c2, _ := excelize.CoordinatesToCellName(xC, xR+1)
		cls, _ := xf.GetCellValue(sheet, c1)
		key, _ := xf.GetCellValue(sheet, c2)
		//fmt.Println("Cell1:", c1, " Cell2:", c2, "--- ClusterName:", cls, "Alert Name:", key)
		if _, ok := data[cls][key]; ok {
			data[cls][key] = data[cls][key] + 1
		} else {
			for i := 0; i < len(cfg.Clusters); i++ {
				if !cfg.Clusters[i].Enable {
					continue
				}
				if len(data[cfg.Clusters[i].Name]) == 0 {
					data2 := make(map[string]int)
					data2[key] = 0
					data[cfg.Clusters[i].Name] = data2
				} else {
					data[cfg.Clusters[i].Name][key] = 0
				}
			}
			csvHead = append(csvHead, key)
			keys = append(keys, key)
			data[cls][key] = data[cls][key] + 1
		}
	}
	//fmt.Println(data)

	xf.SetSheetRow("Summary", cell, &csvHead)
	xC, xR, _ := excelize.CellNameToCoordinates(cell)

	for cls := range data {
		csvData = nil
		csvData = append(csvData, cls)
		for _, key := range keys {
			csvData = append(csvData, data[cls][key])
		}
		c, _ := excelize.CoordinatesToCellName(xC, xR+1)
		xf.SetSheetRow("Summary", c, &csvData)
		xR++
	}
	autoFit(xf, "Summary", 0.9, 13)
}

func uploadFileS3(filename string) {
	bucket := aws.String(cfg.Output.S3.Bucket)
	key := aws.String(filename)
	s3Config := aws.NewConfig()

	if strings.EqualFold(cfg.Output.S3.Provider, "Minio") {
		s3Config = &aws.Config{
			Credentials:      credentials.NewStaticCredentials(cfg.Output.S3.AccessKeyID, cfg.Output.S3.SecretAccessKey, ""),
			Endpoint:         aws.String(cfg.Output.S3.Endpoint),
			Region:           aws.String(cfg.Output.S3.Region),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		}
	} else if strings.EqualFold(cfg.Output.S3.Provider, "AWS") {
		s3Config = &aws.Config{
			Credentials: credentials.NewStaticCredentials(cfg.Output.S3.AccessKeyID, cfg.Output.S3.SecretAccessKey, ""),
			Region:      aws.String(cfg.Output.S3.Region),
		}
	} else {
		erro.Println("AWS or Minio are the only S3 providers right now. Please specify one of them")
	}

	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)

	file, err := os.Open(filename)
	if err != nil {
		erro.Printf("Report file %s cannot be opened", filename)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Body:          fileBytes,
		Bucket:        bucket,
		Key:           key,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	})
	if err != nil {
		erro.Printf("Failed to upload data to %s/%s, %s", *bucket, *key, err.Error())
	} else {
		info.Printf("Report file %s successfully uploaded to bucket %s with key %s", filename, *bucket, *key)
	}
}
