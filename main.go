package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"gopkg.in/yaml.v2"
)

var currentTime time.Time = time.Now()

var (
	xf               *excelize.File = excelize.NewFile() // Output Excel File
	xsh0, xsr1, xsr2 int                                 // Excel Styles for table headers and rows
	ff               string                              // Flag for input config file (YAML)
)

func init() {
	debu = log.New(os.Stdout, "[D] ", log.Ldate|log.Ltime)
	info = log.New(os.Stdout, "[I] ", log.Ldate|log.Ltime)
	warn = log.New(os.Stdout, "[W] ", log.Ldate|log.Ltime)
	erro = log.New(os.Stdout, "[E] ", log.Ldate|log.Ltime)

	var printVersion bool
	flag.BoolVar(&printVersion, "v", false, "Prints version")
	flag.StringVar(&ff, "f", "ocinfo.yaml", "Sets YAML file to configure OCinfo")
	flag.Parse()

	if printVersion {
		info.Printf("OCinfo Version: %s", version)
		info.Printf("Github Repo   : https://github.com/orcunuso/ocinfo")
		os.Exit(0)
	}

	configBytes, err := ioutil.ReadFile(ff)
	perror(err, true)
	err = yaml.Unmarshal(configBytes, &cfg)
	perror(err, true)

	err = validateConfig()
	if err != nil {
		os.Exit(101)
	}

	xsh0, _ = xf.NewStyle(`{
		"font": {"family": "Consolas", "size": 8, "color": "#FFFFFF", "bold": true},
		"border": [
			{"type": "left", "color": "000000", "style": 1}, 
			{"type": "right", "color": "000000", "style": 1}, 
			{"type": "top", "color": "000000", "style": 1}, 
			{"type": "bottom", "color": "000000", "style": 1}],
		"fill": {"type": "pattern","color": ["#000000"],"pattern": 1}}`)
	xsr1, _ = xf.NewStyle(`{
		"font": {"family": "Consolas", "size": 8, "color": "#000000", "bold": false},
		"border": [
			{"type": "left", "color": "000000", "style": 1}, 
			{"type": "right", "color": "000000", "style": 1}, 
			{"type": "top", "color": "000000", "style": 1}, 
			{"type": "bottom", "color": "000000", "style": 1}],
		"fill": {"type": "pattern","color": ["#DCDCDC"],"pattern": 1}}`)
	xsr2, _ = xf.NewStyle(`{
		"font": {"family": "Consolas", "size": 8, "color": "#000000", "bold": false},
		"border": [
			{"type": "left", "color": "000000", "style": 1}, 
			{"type": "right", "color": "000000", "style": 1}, 
			{"type": "top", "color": "000000", "style": 1}, 
			{"type": "bottom", "color": "000000", "style": 1}],
		"fill": {"type": "pattern","color": ["#FFFFFF"],"pattern": 1}}`)
	xf.SetSheetName("Sheet1", "Summary")
}

func main() {
	info.Println("====================== Gathering data from clusters ========================")

	// Create excel sheets seperated for each resource types
	createClusterSheet()
	if cfg.Sheets.Alerts {
		createAlertSheet()
	}
	if cfg.Sheets.Nodes {
		createNodeSheet()
	}
	if cfg.Sheets.Machines {
		createMachineSheet()
	}
	if cfg.Sheets.Namespaces {
		createNamespaceSheet()
	}
	if cfg.Sheets.Nsquotas {
		createQuotaSheet()
	}
	if cfg.Sheets.Pvolumes {
		createPVolumeSheet()
	}
	if cfg.Sheets.Daemonsets {
		createDaemonsetSheet()
	}
	if cfg.Sheets.Routes {
		createRouteSheet()
	}
	if cfg.Sheets.Services {
		createServiceSheet()
	}
	if cfg.Sheets.Deployments {
		createDeploymentSheet()
	}
	if cfg.Sheets.Pods {
		createPodSheet()
	}
	createSummarySheet()

	xf.SetActiveSheet(xf.GetSheetIndex("Summary"))

	// Save excel file and upload to S3 bucket
	info.Println("============================== Exporting data ==============================")
	fileName := "ocinfo_" + currentTime.Format("20060102") + ".xlsx"
	err := xf.SaveAs(fileName)
	if err != nil {
		erro.Println("Report cannot be created. Reason: ", err)
		os.Exit(1)
	}
	info.Println("Report exported as", fileName)

	if cfg.Output.S3.Provider != "" {
		uploadFileS3(fileName)
	}

	info.Println("OCinfo successfully terminated")

}
