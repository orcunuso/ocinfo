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

const version string = "v1.0.0"

var currentTime time.Time = time.Now()

var (
	xf               *excelize.File = excelize.NewFile() // Excel File
	xsh0, xsr1, xsr2 int                                 // Excel Styles for table headers and rows
	ff               string                              // Flag for input config file (YAML)
	cfg              Configuration                       // Configurations get from YAML file
)

// Configuration struct to keep OCinfo properties extracted from YAML configuration file.
type Configuration struct {
	Clusters []struct {
		Name      string `yaml:"name"`
		Enable    bool   `yaml:"enable"`
		BaseURL   string `yaml:"baseURL"`
		Token     string `yaml:"token"`
		PromToken string `yaml:"promToken,omitempty"`
		Quota     string `yaml:"quota"`
	} `yaml:"clusters"`
	Sheets struct {
		Alerts     bool `yaml:"alerts"`
		Namespaces bool `yaml:"namespaces"`
		Nodes      bool `yaml:"nodes"`
		Machines   bool `yaml:"machines"`
		Nsquotas   bool `yaml:"nsquotas"`
		Services   bool `yaml:"services"`
		Routes     bool `yaml:"routes"`
		Pvolumes   bool `yaml:"pvolumes"`
	} `yaml:"sheets"`
}

func init() {
	info = log.New(os.Stdout, "[I] ", log.Ldate|log.Ltime)
	warn = log.New(os.Stdout, "[W] ", log.Ldate|log.Ltime)
	erro = log.New(os.Stdout, "[E] ", log.Ldate|log.Ltime)

	var boolVersion bool
	flag.BoolVar(&boolVersion, "v", false, "Prints version")
	flag.StringVar(&ff, "f", "ocinfo.yaml", "Set YAML file to configure OCinfo")
	flag.Parse()

	if boolVersion {
		info.Printf(magenta("OCinfo Version: ", version))
		os.Exit(0)
	}

	configBytes, err := ioutil.ReadFile(ff)
	perror(err)
	err = yaml.Unmarshal(configBytes, &cfg)
	perror(err)

	xsh0, _ = xf.NewStyle(`{
		"font": {"family": "Calibri", "size": 10, "color": "#FFFFFF", "bold": true},
		"border": [
			{"type": "left", "color": "000000", "style": 1}, 
			{"type": "right", "color": "000000", "style": 1}, 
			{"type": "top", "color": "000000", "style": 1}, 
			{"type": "bottom", "color": "000000", "style": 1}],
		"fill": {"type": "pattern","color": ["#000000"],"pattern": 1}}`)
	xsr1, _ = xf.NewStyle(`{
		"font": {"family": "Calibri", "size": 10, "color": "#000000", "bold": false},
		"border": [
			{"type": "left", "color": "000000", "style": 1}, 
			{"type": "right", "color": "000000", "style": 1}, 
			{"type": "top", "color": "000000", "style": 1}, 
			{"type": "bottom", "color": "000000", "style": 1}],
		"fill": {"type": "pattern","color": ["#DCDCDC"],"pattern": 1}}`)
	xsr2, _ = xf.NewStyle(`{
		"font": {"family": "Calibri", "size": 10, "color": "#000000", "bold": false},
		"border": [
			{"type": "left", "color": "000000", "style": 1}, 
			{"type": "right", "color": "000000", "style": 1}, 
			{"type": "top", "color": "000000", "style": 1}, 
			{"type": "bottom", "color": "000000", "style": 1}],
		"fill": {"type": "pattern","color": ["#FFFFFF"],"pattern": 1}}`)
	xf.SetSheetName("Sheet1", "Summary")
}

func main() {
	// Create excel sheets seperated for each resource types
	info.Printf("OCinfo started with version %s", version)

	getClusters()
	if cfg.Sheets.Alerts {
		getAlerts()
	}
	if cfg.Sheets.Nodes {
		getNodes()
	}
	if cfg.Sheets.Machines {
		getMachines()
	}
	if cfg.Sheets.Namespaces {
		getNamespaces()
	}
	if cfg.Sheets.Nsquotas {
		getNamespaceQuotas()
	}
	if cfg.Sheets.Pvolumes {
		getPVolumes()
	}
	if cfg.Sheets.Services {
		getServices()
	}
	if cfg.Sheets.Routes {
		getRoutes()
	}
	//sheetSummary()

	// Save excel file and quit
	fileName := "ocinfo_" + currentTime.Format("20060102") + ".xlsx"
	err := xf.SaveAs(fileName)
	perror(err)
	info.Printf("OCinfo successfully terminated")
}
