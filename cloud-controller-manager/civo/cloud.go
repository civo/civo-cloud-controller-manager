package civo

import (
	"io"

	"github.com/civo/civogo"
	"k8s.io/client-go/informers"
	cloudprovider "k8s.io/cloud-provider"
)

// ProviderName is the name of the provider.
const ProviderName string = "civo"

// CCMVersion is the version of the CCM.
var CCMVersion string = "dev"

var (
	// APIURL is the URL of the Civo API.
	APIURL string
	// APIKey is the API key for the Civo API.
	APIKey string
	// Region is the region of the Civo API.
	Region string
	// Namespace of the cluster
	Namespace string
	// ClusterID is the ID of the Civo cluster.
	ClusterID string
)

type cloud struct {
	instances     cloudprovider.Instances
	loadbalancers cloudprovider.LoadBalancer
	clients       *clients
}

func init() {
	cloudprovider.RegisterCloudProvider(ProviderName, func(io.Reader) (cloudprovider.Interface, error) {
		return newCloud()
	})
}

func newCloud() (cloudprovider.Interface, error) {
	client, err := civogo.NewClientWithURL(APIKey, APIURL, Region)
	if err != nil {
		return nil, err
	}

	userAgent := &civogo.Component{
		Name:    "civo-cloud-controller-manager",
		ID:      ClusterID,
		Version: CCMVersion,
	}

	client.SetUserAgent(userAgent)

	clients := newClients(client)

	return &cloud{
		clients:       clients,
		instances:     newInstances(clients),
		loadbalancers: newLoadBalancers(clients),
	}, nil
}

// Initialize provides the cloud with a kubernetes client builder and may spawn goroutines
// to perform housekeeping or run custom controllers specific to the cloud provider.
// Any tasks started here should be cleaned up when the stop channel closes.
func (c *cloud) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
	clientset := clientBuilder.ClientOrDie("do-shared-informers")
	sharedInformer := informers.NewSharedInformerFactory(clientset, 0)

	c.clients.kclient = clientset

	sharedInformer.Start(nil)
	sharedInformer.WaitForCacheSync(nil)
}

func (c *cloud) Instances() (cloudprovider.Instances, bool) {
	return c.instances, true
}

func (c *cloud) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return c.loadbalancers, true
}

func (c *cloud) InstancesV2() (cloudprovider.InstancesV2, bool) {
	return nil, false
}

func (c *cloud) Zones() (cloudprovider.Zones, bool) {
	return nil, false
}

func (c *cloud) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

func (c *cloud) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

func (c *cloud) ProviderName() string {
	return ProviderName
}

func (c *cloud) ScrubDNS(nameservers, searches []string) (nsOut, srchOut []string) {
	return nil, nil
}

func (c *cloud) HasClusterID() bool {
	return false
}
