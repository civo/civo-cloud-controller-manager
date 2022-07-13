package main

import (
	"fmt"
	"os"

	"github.com/civo/civo-cloud-controller-manager/cloud-controller-manager/civo"

	"k8s.io/apimachinery/pkg/util/wait"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/cloud-provider/app"
	"k8s.io/cloud-provider/app/config"
	"k8s.io/cloud-provider/options"
	"k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
)

func main() {

	civo.APIURL = os.Getenv("CIVO_API_URL")
	civo.APIKey = os.Getenv("CIVO_API_KEY")
	civo.Region = os.Getenv("CIVO_REGION")
	civo.ClusterID = os.Getenv("CIVO_CLUSTER_ID")

	if civo.APIURL == "" || civo.APIKey == "" || civo.Region == "" || civo.ClusterID == "" {
		fmt.Println("CIVO_API_URL, CIVO_API_KEY, CIVO_REGION, CIVO_CLUSTER_ID environment variables must be set")
		os.Exit(1)
	}

	klog.Infof("Starting ccm with CIVO_API_URL: %s, CIVO_REGION: %s, CIVO_CLUSTER_ID: %s", civo.APIURL, civo.Region, civo.ClusterID)

	opts, err := options.NewCloudControllerManagerOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to construct options: %v\n", err)
		os.Exit(1)
	}

	// Set the provider name
	opts.KubeCloudShared.CloudProvider.Name = civo.ProviderName
	// Disable Cloud Routes
	opts.KubeCloudShared.ConfigureCloudRoutes = false

	command := app.NewCloudControllerManagerCommand(
		opts,
		civoInitializer,
		app.DefaultInitFuncConstructors,
		flag.NamedFlagSets{},
		wait.NeverStop,
	)

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

}

func civoInitializer(cfg *config.CompletedConfig) cloudprovider.Interface {
	cloudConfig := cfg.ComponentConfig.KubeCloudShared.CloudProvider
	// initialize cloud provider with the cloud provider name and config file provided
	cloud, err := cloudprovider.InitCloudProvider(cloudConfig.Name, cloudConfig.CloudConfigFile)
	if err != nil {
		klog.Fatalf("Cloud provider could not be initialized: %v", err)
	}
	if cloud == nil {
		klog.Fatalf("Cloud provider is nil")
	}

	return cloud
}
