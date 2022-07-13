package test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/civo/civo-cloud-controller-manager/cloud-controller-manager/civo"
	"github.com/civo/civogo"
	"github.com/joho/godotenv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/cloud-provider/app"
	"k8s.io/cloud-provider/app/config"
	"k8s.io/cloud-provider/options"
	"k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var CivoRegion, CivoURL string

var e2eTest *E2ETest

type E2ETest struct {
	civo         *civogo.Client
	cluster      *civogo.KubernetesCluster
	tenantClient client.Client
}

// TestMain provisions and then cleans up a cluster for testing against
func TestMain(m *testing.M) {
	e2eTest = &E2ETest{}

	// Recover from a panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
		// Ensure that we clean up the cluster after test tests have run
		e2eTest.cleanUpCluster()
	}()

	// Recover from a SIGINT
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == syscall.SIGINT {
				e2eTest.cleanUpCluster()
			}
		}
	}()

	// Load .env from the project root
	godotenv.Load("../.env")

	// Provision a new cluster
	e2eTest.provisionCluster()
	defer e2eTest.cleanUpCluster()

	// 2. Wait for the cluster to be provisioned
	retry(30, 10*time.Second, func() error {
		cluster, err := e2eTest.civo.GetKubernetesCluster(e2eTest.cluster.ID)
		if err != nil {
			return err
		}
		if cluster.Status != "ACTIVE" {
			return fmt.Errorf("Cluster is not available yet: %s", cluster.Status)
		}
		return nil
	})

	var err error
	e2eTest.cluster, err = e2eTest.civo.GetKubernetesCluster(e2eTest.cluster.ID)
	if err != nil {
		log.Panicf("Unable to fetch ACTIVE Cluster: %s", err.Error())
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(e2eTest.cluster.KubeConfig))
	if err != nil {
		log.Panic(err)
	}

	// Connect to the cluster
	cl, err := client.New(config, client.Options{})
	if err != nil {
		log.Panic(err)
	}
	e2eTest.tenantClient = cl

	deployment := &appsv1.Deployment{}
	retry(5, 5*time.Second, func() error {
		return cl.Get(context.Background(), client.ObjectKey{Namespace: "kube-system", Name: "civo-ccm"}, deployment)
	})

	// 2. Scale down pre-deployed CCM in the cluster
	zero := int32(0)
	deployment.Spec.Replicas = &zero
	err = cl.Update(context.Background(), deployment)
	if err != nil {
		log.Panicf("Unable to scale CCM to zero: %s", err.Error())
	}

	fmt.Println("Running Local Version of CCM")
	secret := &corev1.Secret{}
	err = cl.Get(context.Background(), client.ObjectKey{Namespace: "kube-system", Name: "civo-api-access"}, secret)
	if err != nil {
		log.Panicf("Unable get civo-api-access secret: %s", err.Error())
	}

	// 4. Run the local copy of CCM
	go run(secret, e2eTest.cluster.KubeConfig)
	time.Sleep(1 * time.Minute)

	// 5. Run e2e Tests
	fmt.Println("Running Tests")
	exitCode := m.Run()

	e2eTest.cleanUpCluster()

	os.Exit(exitCode)
}

func (e *E2ETest) provisionCluster() {
	APIKey := os.Getenv("CIVO_API_KEY")
	if APIKey == "" {
		log.Panic("CIVO_API_KEY env variable not provided")
	}

	CivoRegion = os.Getenv("CIVO_REGION")
	if CivoRegion == "" {
		CivoRegion = "LON1"
	}

	CivoURL := os.Getenv("CIVO_URL")
	if CivoURL == "" {
		CivoURL = "https://api.civo.com"
	}

	var err error
	e.civo, err = civogo.NewClientWithURL(APIKey, CivoURL, CivoRegion)
	if err != nil {
		log.Panicf("Unable to initialise Civo Client: %s", err.Error())
	}

	// List Clusters
	list, err := e.civo.ListKubernetesClusters()
	if err != nil {
		log.Panicf("Unable to list Clusters: %s", err.Error())
	}
	for _, cluster := range list.Items {
		if cluster.Name == "ccm-e2e-test" {
			e.cluster = &cluster
			return
		}
	}

	// List Networks
	network, err := e.civo.GetDefaultNetwork()
	if err != nil {
		log.Panicf("Unable to get Default Network: %s", err.Error())
	}

	clusterConfig := &civogo.KubernetesClusterConfig{
		Name:      "ccm-e2e-test",
		Region:    CivoRegion,
		NetworkID: network.ID,
		Pools: []civogo.KubernetesClusterPoolConfig{
			{
				Count: 2,
				Size:  "g4s.kube.xsmall",
			},
		},
	}

	e.cluster, err = e.civo.NewKubernetesClusters(clusterConfig)
	if err != nil {
		log.Panicf("Unable to provision new cluster: %s", err.Error())
	}
}

func (e *E2ETest) cleanUpCluster() {
	fmt.Println("Attempting Test Cleanup")
	if e.cluster != nil {
		fmt.Printf("Deleting Cluster: %s\n", e.cluster.ID)
		// e.civo.DeleteKubernetesCluster(e.cluster.ID)
	}
}

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Println("retrying after error:", err)
			time.Sleep(sleep)
			sleep *= 2
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

func run(secret *corev1.Secret, kubeconfig string) {
	err := os.WriteFile("./kubeconfig", []byte(kubeconfig), 0644)
	if err != nil {
		panic(err)
	}

	// Read env var from in cluster secret
	civo.APIURL = string(secret.Data["api-url"])
	civo.APIKey = string(secret.Data["api-key"])
	civo.Region = string(secret.Data["region"])
	civo.ClusterID = string(secret.Data["cluster-id"])

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
	opts.Kubeconfig = "./kubeconfig"
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
		log.Panicf("Cloud provider could not be initialized: %v", err)
	}
	if cloud == nil {
		log.Panicf("Cloud provider is nil")
	}

	return cloud
}
