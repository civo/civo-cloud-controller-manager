package test

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/civo/civogo"
	"github.com/joho/godotenv"
)

const CivoRegion = "LON1"

type E2ETest struct {
	civo    *civogo.Client
	cluster *civogo.KubernetesCluster
}

// TestMain provisions and then cleans up a cluster for testing against
func TestMain(m *testing.M) {
	e2eTest := &E2ETest{}

	// Recover from a panic
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("There has been a panic")
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

	// 2. Scale down pre-deployed CCM in the cluster
	// 4. Run the local copy of CCM
	// 5. Run e2e Tests

	// Run the tests
	fmt.Println("Running Tests")
	exitCode := m.Run()

	e2eTest.cleanUpCluster()

	os.Exit(exitCode)
}

func (e *E2ETest) provisionCluster() {
	api_key := os.Getenv("CIVO_API_KEY")
	if api_key == "" {
		log.Fatal("CIVO_API_KEY env variable not provided")
	}
	var err error
	e.civo, err = civogo.NewClient(api_key, CivoRegion)
	if err != nil {
		log.Fatalf("Unable to initialise Civo Client: %s", err.Error())
	}

	// List Networks
	network, err := e.civo.GetDefaultNetwork()
	if err != nil {
		log.Fatalf("Unable to get Default Network: %s", err.Error())
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
		log.Fatalf("Unable to provision new cluster: %s", err.Error())
	}
}

func (e *E2ETest) cleanUpCluster() {
	fmt.Println("Attempting Test Cleanup")
	if e.cluster != nil {
		fmt.Printf("Deleting Cluster: %s\n", e.cluster.ID)
		e.civo.DeleteKubernetesCluster(e.cluster.ID)
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
