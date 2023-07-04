package civo

import (
	"context"
	"testing"

	"github.com/civo/civogo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	cloudprovider "k8s.io/cloud-provider"
)

var _ cloudprovider.Instances = new(instances)

// TestNodeAddresses tests the NodeAddresses method
func TestNodeAddresses(t *testing.T) {
	clusterID := "26b704b7-2c90-4378-b9de-046b5d91ace6"
	instanceStore := []civogo.Instance{
		{
			ID:        "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
			Hostname:  "test-node",
			PrivateIP: "192.168.11.11",
		},
	}
	expected :=
		[]corev1.NodeAddress{
			{
				Type:    corev1.NodeInternalIP,
				Address: "192.168.11.11",
			},
		}
	clusterStore := []civogo.KubernetesCluster{
		{
			ID: "26b704b7-2c90-4378-b9de-046b5d91ace6",
			Instances: []civogo.KubernetesInstance{
				{
					ID:       "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
					Hostname: "test-node",
					PublicIP: "192.168.11.11",
				},
			},
		},
	}

	g := NewWithT(t)

	fakeCivoClient, err := civogo.NewFakeClient()
	g.Expect(err).To(BeNil())

	fakeCivoClient.Instances = instanceStore
	fakeCivoClient.Clusters = clusterStore

	ClusterID = clusterID

	fakeK8s := fake.NewSimpleClientset()
	clients := &clients{
		civoClient: fakeCivoClient,
		kclient:    fakeK8s,
	}

	instances := &instances{
		client: clients,
	}

	addresses, err := instances.NodeAddresses(context.TODO(), "test-node")
	g.Expect(err).To(BeNil())
	g.Expect(addresses).To(Equal(expected))
}

// TestNodeAddressesByProviderID tests the NodeAddressesByProviderID method
func TestNodeAddressesByProviderID(t *testing.T) {
	clusterID := "26b704b7-2c90-4378-b9de-046b5d91ace6"
	providerID := "civo://acb5cbd0-ef7f-4edd-8a51-005940fdb7a8"
	instanceStore := []civogo.Instance{
		{
			ID:        "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
			Hostname:  "test-node",
			PrivateIP: "192.168.11.11",
			PublicIP:  "127.0.0.101",
		},
	}
	expected :=
		[]corev1.NodeAddress{
			{
				Type:    corev1.NodeInternalIP,
				Address: "192.168.11.11",
			},
			corev1.NodeAddress{
				Type:    corev1.NodeExternalIP,
				Address: "127.0.0.101",
			},
		}
	clusterStore := []civogo.KubernetesCluster{
		{
			ID: "26b704b7-2c90-4378-b9de-046b5d91ace6",
			Instances: []civogo.KubernetesInstance{
				{
					ID:       "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
					Hostname: "test-node",
					PublicIP: "192.168.11.11",
				},
			},
		},
	}

	g := NewWithT(t)

	fakeCivoClient, err := civogo.NewFakeClient()
	g.Expect(err).To(BeNil())

	fakeCivoClient.Instances = instanceStore
	fakeCivoClient.Clusters = clusterStore

	ClusterID = clusterID

	fakeK8s := fake.NewSimpleClientset()
	clients := &clients{
		civoClient: fakeCivoClient,
		kclient:    fakeK8s,
	}

	instances := &instances{
		client: clients,
	}

	addresses, err := instances.NodeAddressesByProviderID(context.TODO(), providerID)
	g.Expect(err).To(BeNil())
	g.Expect(addresses).To(Equal(expected))
}

// TestInstanceID tests the InstanceID method
func TestInstanceID(t *testing.T) {
	clusterID := "26b704b7-2c90-4378-b9de-046b5d91ace6"
	instanceStore := []civogo.Instance{
		{
			ID:        "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
			Hostname:  "test-node",
			PrivateIP: "192.168.11.11",
		},
	}
	expected := "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8"
	clusterStore := []civogo.KubernetesCluster{
		{
			ID: "26b704b7-2c90-4378-b9de-046b5d91ace6",
			Instances: []civogo.KubernetesInstance{
				{
					ID:       "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
					Hostname: "test-node",
					PublicIP: "192.168.11.11",
				},
			},
		},
	}

	g := NewWithT(t)

	fakeCivoClient, err := civogo.NewFakeClient()
	g.Expect(err).To(BeNil())

	fakeCivoClient.Instances = instanceStore
	fakeCivoClient.Clusters = clusterStore

	ClusterID = clusterID

	fakeK8s := fake.NewSimpleClientset()
	clients := &clients{
		civoClient: fakeCivoClient,
		kclient:    fakeK8s,
	}

	instances := &instances{
		client: clients,
	}

	id, err := instances.InstanceID(context.TODO(), "test-node")
	g.Expect(err).To(BeNil())
	g.Expect(id).To(Equal(expected))
}

// TestInstanceType tests the InstanceID method
func TestInstanceType(t *testing.T) {
	clusterID := "26b704b7-2c90-4378-b9de-046b5d91ace6"
	instanceStore := []civogo.Instance{
		{
			ID:        "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
			Hostname:  "test-node",
			Size:      "g3.k3s.small",
			PrivateIP: "192.168.11.11",
		},
	}
	expected := "g3.k3s.small"
	clusterStore := []civogo.KubernetesCluster{
		{
			ID: "26b704b7-2c90-4378-b9de-046b5d91ace6",
			Instances: []civogo.KubernetesInstance{
				{
					ID:       "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
					Hostname: "test-node",
					Size:     "g3.k3s.small",
					PublicIP: "192.168.11.11",
				},
			},
		},
	}

	g := NewWithT(t)

	fakeCivoClient, err := civogo.NewFakeClient()
	g.Expect(err).To(BeNil())

	fakeCivoClient.Instances = instanceStore
	fakeCivoClient.Clusters = clusterStore

	ClusterID = clusterID

	fakeK8s := fake.NewSimpleClientset()
	clients := &clients{
		civoClient: fakeCivoClient,
		kclient:    fakeK8s,
	}

	instances := &instances{
		client: clients,
	}

	id, err := instances.InstanceType(context.TODO(), "test-node")
	g.Expect(err).To(BeNil())
	g.Expect(id).To(Equal(expected))
}

// TestInstanceTypeByProviderID tests the InstanceID method
func TestInstanceTypeByProviderID(t *testing.T) {
	clusterID := "26b704b7-2c90-4378-b9de-046b5d91ace6"
	providerID := "civo://acb5cbd0-ef7f-4edd-8a51-005940fdb7a8"
	instanceStore := []civogo.Instance{
		{
			ID:        "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
			Hostname:  "test-node",
			Size:      "g3.k3s.small",
			PrivateIP: "192.168.11.11",
		},
	}
	expected := "g3.k3s.small"
	clusterStore := []civogo.KubernetesCluster{
		{
			ID: "26b704b7-2c90-4378-b9de-046b5d91ace6",
			Instances: []civogo.KubernetesInstance{
				{
					ID:       "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
					Hostname: "test-node",
					Size:     "g3.k3s.small",
					PublicIP: "192.168.11.11",
				},
			},
		},
	}

	g := NewWithT(t)

	fakeCivoClient, err := civogo.NewFakeClient()
	g.Expect(err).To(BeNil())

	fakeCivoClient.Instances = instanceStore
	fakeCivoClient.Clusters = clusterStore

	ClusterID = clusterID

	fakeK8s := fake.NewSimpleClientset()
	clients := &clients{
		civoClient: fakeCivoClient,
		kclient:    fakeK8s,
	}

	instances := &instances{
		client: clients,
	}

	id, err := instances.InstanceTypeByProviderID(context.TODO(), providerID)
	g.Expect(err).To(BeNil())
	g.Expect(id).To(Equal(expected))
}

// TestAddSSHKeyToAllInstances tests the AddSSHKeyToAllInstances method
func TestAddSSHKeyToAllInstances(t *testing.T) {
	g := NewWithT(t)

	fakeCivoClient, err := civogo.NewFakeClient()
	g.Expect(err).To(BeNil())

	fakeK8s := fake.NewSimpleClientset()
	clients := &clients{
		civoClient: fakeCivoClient,
		kclient:    fakeK8s,
	}

	instances := &instances{
		client: clients,
	}

	err = instances.AddSSHKeyToAllInstances(context.TODO(), "user", []byte("test-key"))
	g.Expect(err).To(BeNil())
}

// TestCurrentNodeName tests the CurrentNodeName method
func TestCurrentNodeName(t *testing.T) {
	g := NewWithT(t)

	fakeCivoClient, err := civogo.NewFakeClient()
	g.Expect(err).To(BeNil())

	fakeK8s := fake.NewSimpleClientset()
	clients := &clients{
		civoClient: fakeCivoClient,
		kclient:    fakeK8s,
	}

	instances := &instances{
		client: clients,
	}

	name, err := instances.CurrentNodeName(context.TODO(), "test-node")
	g.Expect(err).To(BeNil())
	g.Expect(name).To(Equal(types.NodeName("test-node")))
}

// TestInstanceExistsByProviderID tests the InstanceExistsByProviderID method
func TestInstanceExistsByProviderID(t *testing.T) {
	clusterID := "26b704b7-2c90-4378-b9de-046b5d91ace6"
	providerID := "civo://acb5cbd0-ef7f-4edd-8a51-005940fdb7a8"
	instanceStore := []civogo.Instance{
		{
			ID:        "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
			Hostname:  "test-node",
			Size:      "g3.k3s.small",
			PrivateIP: "192.168.11.11",
		},
	}
	clusterStore := []civogo.KubernetesCluster{
		{
			ID: "26b704b7-2c90-4378-b9de-046b5d91ace6",
			Instances: []civogo.KubernetesInstance{
				{
					ID:       "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
					Hostname: "test-node",
					Size:     "g3.k3s.small",
					PublicIP: "192.168.11.11",
				},
			},
		},
	}

	g := NewWithT(t)

	fakeCivoClient, err := civogo.NewFakeClient()
	g.Expect(err).To(BeNil())

	fakeCivoClient.Instances = instanceStore
	fakeCivoClient.Clusters = clusterStore

	ClusterID = clusterID

	fakeK8s := fake.NewSimpleClientset()
	clients := &clients{
		civoClient: fakeCivoClient,
		kclient:    fakeK8s,
	}

	instances := &instances{
		client: clients,
	}

	exists, err := instances.InstanceExistsByProviderID(context.TODO(), providerID)
	g.Expect(err).To(BeNil())
	g.Expect(exists).To(BeTrue())

	exists, err = instances.InstanceExistsByProviderID(context.TODO(), "civo://non-existing-provider-id")
	g.Expect(err).To(Not(BeNil()))
	g.Expect(exists).To(BeTrue())
}

// TestInstanceShutdownByProviderID tests the InstanceShutdownByProviderID method
func TestInstanceShutdownByProviderID(t *testing.T) {
	clusterID := "26b704b7-2c90-4378-b9de-046b5d91ace6"
	providerID := "civo://acb5cbd0-ef7f-4edd-8a51-005940fdb7a8"
	instanceStore := []civogo.Instance{
		{
			ID:        "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
			Hostname:  "test-node",
			Size:      "g3.k3s.small",
			PrivateIP: "192.168.11.11",
		},
	}
	clusterStore := []civogo.KubernetesCluster{
		{
			ID: "26b704b7-2c90-4378-b9de-046b5d91ace6",
			Instances: []civogo.KubernetesInstance{
				{
					ID:       "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8",
					Hostname: "test-node",
					Size:     "g3.k3s.small",
					PublicIP: "192.168.11.11",
				},
			},
		},
	}

	g := NewWithT(t)

	fakeCivoClient, err := civogo.NewFakeClient()
	g.Expect(err).To(BeNil())

	fakeCivoClient.Instances = instanceStore
	fakeCivoClient.Clusters = clusterStore

	ClusterID = clusterID

	fakeK8s := fake.NewSimpleClientset()
	clients := &clients{
		civoClient: fakeCivoClient,
		kclient:    fakeK8s,
	}

	instances := &instances{
		client: clients,
	}

	exists, err := instances.InstanceShutdownByProviderID(context.TODO(), providerID)
	g.Expect(err).To(BeNil())
	g.Expect(exists).To(BeTrue())

	exists, err = instances.InstanceShutdownByProviderID(context.TODO(), "civo://non-existing-provider-id")
	g.Expect(err).To(BeNil())
	g.Expect(exists).To(BeFalse())
}
