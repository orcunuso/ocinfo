package main

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

const version string = "v0.2.1"

var (
	testedOCPVersions  = map[string]bool{"4.5": true, "4.6": true}
	testedOCPProviders = map[string]bool{"oVirt": true, "BareMetal": true}
)

var cfg Configuration // Configuration items from YAML file

// Configuration struct to keep OCinfo properties extracted from YAML configuration file.
type Configuration struct {
	Clusters []struct {
		Name      string `yaml:"name"`
		Version   string `yaml:"version,omitempty"`
		Provider  string `yaml:"provider,omitempty"`
		Enable    bool   `yaml:"enable"`
		BaseURL   string `yaml:"baseURL"`
		Token     string `yaml:"token"`
		PromToken string `yaml:"promToken,omitempty"`
		Quota     string `yaml:"quota"`
	} `yaml:"clusters"`
	Sheets struct {
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

func validateConfig() error {
	validated := true
	info.Printf("OCinfo started with version %s", version)
	info.Println("======================== Validating configuration =========================")

	// Check Clusters
	validatedClusters := 0
	for i := 0; i < len(cfg.Clusters); i++ {
		if cfg.Clusters[i].Name == "" || cfg.Clusters[i].Token == "" {
			warn.Printf(yellow(cfg.Clusters[i].Name, " -> Cluster name or token can't be empty. Omitting."))
			continue
		}
		if cfg.Clusters[i].Enable == false {
			info.Printf("%s -> Cluster is not enabled in the configuration.", cfg.Clusters[i].Name)
			continue
		}
		if strings.Index(cfg.Clusters[i].BaseURL, "https://") != 0 {
			warn.Printf(yellow(cfg.Clusters[i].Name, " -> Cluster API URL does not start with https://. Omitting."))
			continue
		}

		apiurl := cfg.Clusters[i].BaseURL + "/apis/config.openshift.io/v1/clusterversions?limit=100"
		body, _ := getRest(apiurl, cfg.Clusters[i].Token)
		items := gjson.GetBytes(body, "items")
		items.ForEach(func(key, value gjson.Result) bool {
			vars := gjson.Get(value.String(), `spec.desiredUpdate.version`)
			cfg.Clusters[i].Version = strings.Split(vars.String(), ".")[0] + "." + strings.Split(vars.String(), ".")[1]
			return true
		})

		if cfg.Clusters[i].Version != "" {
			info.Printf("%s -> Cluster is running with version %s\n", cfg.Clusters[i].Name, cfg.Clusters[i].Version)
		} else {
			warn.Printf(yellow(cfg.Clusters[i].Name, " -> Can't get cluster version. Something is wrong."))
			continue
		}

		if _, ok := testedOCPVersions[cfg.Clusters[i].Version]; !ok {
			warn.Printf(yellow(cfg.Clusters[i].Name, " -> OCinfo is not well tested for this version. You may expect empty values or inconsistent behaviors"))
		}

		validatedClusters++
	}
	info.Printf("There are %v cluster(s) properly configured\n", validatedClusters)
	if validatedClusters < 1 {
		validated = false
	}

	// Check Output.S3 configuration
	if cfg.Output.S3.Provider == "" {
		info.Println("No S3 provider defined. An excel report will be created on working directory.")
	} else {
		if strings.EqualFold(cfg.Output.S3.Provider, "AWS") {
			if cfg.Output.S3.Region == "" || cfg.Output.S3.Bucket == "" || cfg.Output.S3.AccessKeyID == "" || cfg.Output.S3.SecretAccessKey == "" {
				warn.Println(yellow("S3.AWS configuration is not complete. Please check region, bucket, accessKeyID and secretAccessKey values."))
				validated = false
			}
		} else if strings.EqualFold(cfg.Output.S3.Provider, "Minio") {
			if cfg.Output.S3.Endpoint == "" || cfg.Output.S3.Bucket == "" || cfg.Output.S3.AccessKeyID == "" || cfg.Output.S3.SecretAccessKey == "" {
				warn.Println(yellow("S3.Minio configuration is not complete. Please check endpoint, bucket, accessKeyID and secretAccessKey values."))
				validated = false
			}
		} else {
			warn.Println("AWS and Minio are the only S3 providers right now. Please specify one of them")
			validated = false
		}
	}

	if !validated {
		erro.Println(red("Configuration could not be validated."))
		info.Println("Please refer to https://github.com/orcunuso/ocinfo/blob/main/README.md for more information")
		return fmt.Errorf("configuration could not be validated")
	}
	return nil
}
