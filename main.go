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
	Appnslabel string `yaml:"appnslabel"`
	Sheets     struct {
		Alerts      bool `yaml:"alerts,omitempty"`
		Namespaces  bool `yaml:"namespaces,omitempty"`
		Nodes       bool `yaml:"nodes,omitempty"`
		Machines    bool `yaml:"machines,omitempty"`
		Nsquotas    bool `yaml:"nsquotas,omitempty"`
		Services    bool `yaml:"services,omitempty"`
		Routes      bool `yaml:"routes,omitempty"`
		Pvolumes    bool `yaml:"pvolumes,omitempty"`
		Daemonsets  bool `yaml:"daemonsets,omitempty"`
		Pods        bool `yaml:"pods,omitempty"`
		Deployments bool `yaml:"deployments,omitempty"`
	} `yaml:"sheets,omitempty"`
	Output struct {
		S3 struct {
			Provider        string `yaml:"provider,omitempty"`
			Endpoint        string `yaml:"endpoint,omitempty"`
			Region          string `yaml:"region,omitempty"`
			Bucket          string `yaml:"bucket,omitempty"`
			AccessKeyID     string `yaml:"accessKeyID,omitempty"`
			SecretAccessKey string `yaml:"secretAccessKey,omitempty"`
		} `yaml:"s3"`
	} `yaml:"output"`
}

func init() {
	info = log.New(os.Stdout, "[I] ", log.Ldate|log.Ltime)
	warn = log.New(os.Stdout, "[W] ", log.Ldate|log.Ltime)
	erro = log.New(os.Stdout, "[E] ", log.Ldate|log.Ltime)

	var boolVersion bool
	flag.BoolVar(&boolVersion, "v", false, "Prints version")
	flag.StringVar(&ff, "f", "ocinfo.yaml", "Sets YAML file to configure OCinfo")
	flag.Parse()

	if boolVersion {
		info.Printf(magenta("OCinfo Version: ", version))
		os.Exit(0)
	}

	configBytes, err := ioutil.ReadFile(ff)
	perror(err, true)
	err = yaml.Unmarshal(configBytes, &cfg)
	perror(err, true)

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
	// Create excel sheets seperated for each resource types
	info.Printf("OCinfo started with version %s", version)

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
