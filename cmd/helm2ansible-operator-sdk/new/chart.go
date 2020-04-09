package new

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/redhat-nfvpe/helm2ansible-operator-sdk/internal/pathconfig"
	"github.com/redhat-nfvpe/helm2ansible-operator-sdk/internal/pkg/templating"
	"github.com/redhat-nfvpe/helm2ansible-operator-sdk/internal/resourcecache"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
)

//HelmChartClient ....
type HelmChartClient struct {
	HelmChartRef      string
	HelmChartRepo     string
	HelmChartVersion  string
	HelmChartCAFile   string
	HelmChartCertFile string
	HelmChartKeyFile  string
	Username          string
	Password          string
	Chart             *chart.Chart
	ChartName         string
	PathConfig        *pathconfig.PathConfig
}

//NewChartClient creates a new chart client
func NewChartClient() *HelmChartClient {
	client := HelmChartClient{}
	client.PathConfig, _ = GetBasePathConfig()
	return &client
}

// SetValues ingests all the necessary values for the client
func (hc *HelmChartClient) SetValues(helmChartRef, helmChartVersion, helmChartRepo, username, password, helmChartCAFile, helmChartCertFile, helmChartKeyFile string) {
	hc.HelmChartRef = helmChartRef
	hc.HelmChartVersion = helmChartVersion
	hc.HelmChartRepo = helmChartRepo
	hc.Username = username
	hc.Password = password
	hc.HelmChartCAFile = helmChartCAFile
	hc.HelmChartCertFile = helmChartCertFile
	hc.HelmChartKeyFile = helmChartKeyFile
}

//LoadChart uses the chart client's values to retrieve  the appropriate chart
func (hc *HelmChartClient) LoadChart() error {
	var chartPath string
	chartPath = hc.HelmChartRef
	if len(hc.HelmChartRepo) > 0 {

		var out io.Writer
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		d := downloader.ChartDownloader{
			Out:      out,
			Verify:   downloader.VerifyNever,
			Keyring:  "",
			HelmHome: helmpath.Home(filepath.Join(home, ".helm")),
			Getters:  getter.All(environment.EnvSettings{}),
			Username: hc.Username,
			Password: hc.Password,
		}

		chartURL, err := repo.FindChartInAuthRepoURL(hc.HelmChartRepo, hc.Username, hc.Password, chartPath, hc.HelmChartVersion, hc.HelmChartCertFile, hc.HelmChartKeyFile, hc.HelmChartCAFile, getter.All(environment.EnvSettings{}))
		if err != nil {
			return err
		}

		hc.HelmChartRef = chartURL

		downloaded, _, err := d.DownloadTo(hc.HelmChartRef, hc.HelmChartVersion, hc.PathConfig.GetBasePath())
		if err != nil {
			log.Printf("Errored here")
			return err
		}

		chartutil.ExpandFile(hc.PathConfig.GetBasePath(), downloaded)
		os.RemoveAll(downloaded)
		log.Printf("Downloaded Chart To: %v\n", hc.PathConfig.GetBasePath())
		chartPath = filepath.Join(hc.PathConfig.GetBasePath(), chartPath)
	}

	loadedChart, err := chartutil.Load(chartPath)
	if err != nil {
		return err
	}
	hc.Chart = loadedChart
	hc.ChartName = hc.Chart.Metadata.GetName()

	return nil
}
func (hc *HelmChartClient) CopyTemplates(dst string) ([]string, error) {
	var copiedFiles []string
	for _, template := range hc.Chart.Templates {
		fileName := strings.Split(template.GetName(), "/")
		if filepath.Ext(fileName[1]) == ".yaml" || filepath.Ext(fileName[1]) == ".yml" || filepath.Ext(fileName[1]) == ".tpl" {
			f := fmt.Sprintf("%s", fileName[1]+".j2")
			fmt.Println(f)
			err := ioutil.WriteFile(dst+fileName[1]+".j2", template.GetData(), 0644)
			if err == nil {
				//fmt.Println(template.GetData())
				copiedFiles = append(copiedFiles, f)
			} else {
				fmt.Printf("#%v", err)
			}

		} else {
			fmt.Println(filepath.Ext(fileName[1]))
		}
	}
	return copiedFiles, nil
}

func scaffoldOverwrite(outputDir, kind, apiVersion string, rcache *resourcecache.ResourceCache) error {

	if err := templating.OverwriteController(outputDir, kind, apiVersion, rcache); err != nil {
		return fmt.Errorf("error while overwriting controller: %v", err)
	}
	// create templates for writing to file
	templates := templating.CacheTemplating(rcache, outputDir, kind, apiVersion)
	// templates to files; outputDir is the parent directory where the operator scaffolding lives
	resDir := filepath.Join(outputDir, "pkg", "resources")

	// create the necessary package resource specific folders
	if err := templating.ResourceFileStructure(rcache, resDir); err != nil {
		return fmt.Errorf("error creating resource file structure: %v", err)
	}
	if err := templating.TemplatesToFiles(templates, resDir); err != nil {
		return fmt.Errorf("error writing to template: %v", err)
	}

	return nil
}
