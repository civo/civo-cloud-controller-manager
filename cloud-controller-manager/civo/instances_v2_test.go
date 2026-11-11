package civo

import (
	"context"
	"errors"
	"testing"

	"github.com/civo/civogo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	cloudprovider "k8s.io/cloud-provider"
	cloudproviderapi "k8s.io/cloud-provider/api"
)

var _ cloudprovider.InstancesV2 = new(instancesV2)

const (
	v2ClusterID  = "26b704b7-2c90-4378-b9de-046b5d91ace6"
	v2InstanceID = "acb5cbd0-ef7f-4edd-8a51-005940fdb7a8"
	v2NodeName   = "test-node"
)

var uninitializedTaint = corev1.Taint{
	Key:    cloudproviderapi.TaintExternalCloudProvider,
	Value:  "true",
	Effect: corev1.TaintEffectNoSchedule,
}

func newV2TestSetup(t *testing.T, pools []civogo.KubernetesPool, node *corev1.Node) (*instancesV2, *fake.Clientset) {
	t.Helper()
	g := NewWithT(t)

	fakeCivoClient, err := civogo.NewFakeClient()
	g.Expect(err).To(BeNil())
	fakeCivoClient.Instances = []civogo.Instance{{
		ID:        v2InstanceID,
		Hostname:  v2NodeName,
		PrivateIP: "192.168.11.11",
		Size:      "g4s.kube.small",
		Status:    statusActive,
	}}
	fakeCivoClient.Clusters = []civogo.KubernetesCluster{{
		ID:        v2ClusterID,
		Instances: []civogo.KubernetesInstance{{ID: v2InstanceID, Hostname: v2NodeName}},
		Pools:     pools,
	}}
	ClusterID = v2ClusterID

	fakeK8s := fake.NewSimpleClientset(node)
	return &instancesV2{client: &clients{civoClient: fakeCivoClient, kclient: fakeK8s}}, fakeK8s
}

func newNode(labels map[string]string, taints ...corev1.Taint) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: v2NodeName, Labels: labels},
		Spec:       corev1.NodeSpec{Taints: taints},
	}
}

func TestInstanceMetadataAppliesPoolLabelsAndTaints(t *testing.T) {
	poolTaint := corev1.Taint{Key: "dedicated", Value: "gpu", Effect: corev1.TaintEffectNoSchedule}
	tests := []struct {
		name       string
		pools      []civogo.KubernetesPool
		node       *corev1.Node
		wantLabels map[string]string
		wantTaints []corev1.Taint
	}{
		{
			name: "uninitialized node gets pool labels and taints (matched by instance names)",
			pools: []civogo.KubernetesPool{{
				ID:            "pool-1",
				InstanceNames: []string{v2NodeName},
				Labels:        map[string]string{"pool": "gpu"},
				Taints:        []corev1.Taint{poolTaint},
			}},
			node:       newNode(map[string]string{"existing": "x"}, uninitializedTaint),
			wantLabels: map[string]string{"existing": "x", "pool": "gpu"},
			wantTaints: []corev1.Taint{uninitializedTaint, poolTaint},
		},
		{
			name: "pool matched by instance ID; existing taint value is updated, not duplicated",
			pools: []civogo.KubernetesPool{{
				ID:        "pool-1",
				Instances: []civogo.KubernetesInstance{{ID: v2InstanceID}},
				Taints:    []corev1.Taint{poolTaint},
			}},
			node: newNode(nil, uninitializedTaint,
				corev1.Taint{Key: "dedicated", Value: "old", Effect: corev1.TaintEffectNoSchedule}),
			wantTaints: []corev1.Taint{uninitializedTaint, poolTaint},
		},
		{
			name: "initialized node is left untouched",
			pools: []civogo.KubernetesPool{{
				ID:            "pool-1",
				InstanceNames: []string{v2NodeName},
				Labels:        map[string]string{"pool": "gpu"},
				Taints:        []corev1.Taint{poolTaint},
			}},
			node: newNode(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			i, k := newV2TestSetup(t, tt.pools, tt.node)

			meta, err := i.InstanceMetadata(context.TODO(), tt.node)
			g.Expect(err).To(BeNil())
			g.Expect(meta.ProviderID).To(Equal("civo://" + v2InstanceID))
			g.Expect(meta.InstanceType).To(Equal("g4s.kube.small"))
			g.Expect(meta.NodeAddresses).To(Equal([]corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "192.168.11.11"}}))

			got, err := k.CoreV1().Nodes().Get(context.TODO(), v2NodeName, metav1.GetOptions{})
			g.Expect(err).To(BeNil())
			if tt.wantLabels == nil {
				g.Expect(got.Labels).To(Equal(tt.node.Labels))
			} else {
				g.Expect(got.Labels).To(Equal(tt.wantLabels))
			}
			if tt.wantTaints == nil {
				g.Expect(got.Spec.Taints).To(Equal(tt.node.Spec.Taints))
			} else {
				g.Expect(got.Spec.Taints).To(Equal(tt.wantTaints))
			}
		})
	}
}

func TestInstanceMetadataFailsWhenPoolNotFound(t *testing.T) {
	g := NewWithT(t)
	node := newNode(nil, uninitializedTaint)
	i, _ := newV2TestSetup(t, []civogo.KubernetesPool{{ID: "other", InstanceNames: []string{"other-node"}}}, node)

	_, err := i.InstanceMetadata(context.TODO(), node)
	g.Expect(err).To(HaveOccurred())
}

func TestInstanceMetadataFailsWhenClusterLookupFails(t *testing.T) {
	g := NewWithT(t)
	node := newNode(nil, uninitializedTaint)
	i, _ := newV2TestSetup(t, nil, node)
	fc := i.client.civoClient.(*civogo.FakeClient)
	i.client.civoClient = &clusterErrorClient{FakeClient: fc, err: errors.New("api down")}

	_, err := i.InstanceMetadata(context.TODO(), node)
	g.Expect(err).To(HaveOccurred())
}

func TestInstanceExistsV2(t *testing.T) {
	g := NewWithT(t)
	i, _ := newV2TestSetup(t, nil, newNode(nil))

	exists, err := i.InstanceExists(context.TODO(), newNode(nil))
	g.Expect(err).To(BeNil())
	g.Expect(exists).To(BeTrue())

	gone := newNode(nil)
	gone.Spec.ProviderID = "civo://does-not-exist"
	exists, err = i.InstanceExists(context.TODO(), gone)
	g.Expect(err).To(BeNil())
	g.Expect(exists).To(BeFalse())
}

func TestInstanceShutdownV2(t *testing.T) {
	g := NewWithT(t)
	i, _ := newV2TestSetup(t, nil, newNode(nil))

	shutdown, err := i.InstanceShutdown(context.TODO(), newNode(nil))
	g.Expect(err).To(BeNil())
	g.Expect(shutdown).To(BeFalse())
}
