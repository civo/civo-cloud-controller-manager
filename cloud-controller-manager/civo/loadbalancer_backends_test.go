package civo

import (
	"context"
	"testing"

	"github.com/civo/civogo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestBuildBackends(t *testing.T) {
	g := NewWithT(t)

	service := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Protocol: corev1.ProtocolTCP, Port: 80, NodePort: 30080},
			},
		},
	}

	nodes := []*corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "hostname-first"},
			Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{
				{Type: corev1.NodeHostName, Address: "hostname-first"},
				{Type: corev1.NodeInternalIP, Address: "192.168.1.1"},
			}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "external-first"},
			Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{
				{Type: corev1.NodeExternalIP, Address: "203.0.113.2"},
				{Type: corev1.NodeInternalIP, Address: "192.168.1.2"},
			}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "not-initialized"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "untyped"},
			Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{
				{Address: "192.168.1.4"},
			}},
		},
	}

	var backends []civogo.LoadBalancerBackendConfig
	g.Expect(func() { backends = buildBackends(service, nodes) }).NotTo(Panic())

	ips := []string{}
	for _, b := range backends {
		ips = append(ips, b.IP)
		g.Expect(b.SourcePort).To(Equal(int32(80)))
		g.Expect(b.TargetPort).To(Equal(int32(30080)))
	}
	g.Expect(ips).To(Equal([]string{"192.168.1.1", "192.168.1.2", "192.168.1.4"}))
}

// TestEnsureLoadBalancerSkipsNodeWithoutAddress is a regression test for a
// panic (index out of range) when one of the nodes passed by the service
// controller had no address yet.
func TestEnsureLoadBalancerSkipsNodeWithoutAddress(t *testing.T) {
	g := NewWithT(t)

	fakeCivoClient, err := civogo.NewFakeClient()
	g.Expect(err).To(BeNil())
	fakeCivoClient.Clusters = []civogo.KubernetesCluster{
		{ID: "a32fe5eb-1922-43e8-81bc-7f83b4011334", Name: "test"},
	}
	ClusterID = fakeCivoClient.Clusters[0].ID

	clients := newClients(fakeCivoClient)
	clients.kclient = fake.NewSimpleClientset()

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: corev1.NamespaceDefault},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{
				{Name: "http", Protocol: corev1.ProtocolTCP, Port: 80, NodePort: 30000},
			},
		},
	}
	_, err = clients.kclient.CoreV1().Services(corev1.NamespaceDefault).Create(context.Background(), service, metav1.CreateOptions{})
	g.Expect(err).To(BeNil())

	nodes := []*corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node1"},
			Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{
				{Type: corev1.NodeInternalIP, Address: "192.168.1.1"},
			}},
		},
		{ObjectMeta: metav1.ObjectMeta{Name: "node2"}},
	}

	lb := &loadbalancer{client: clients}
	g.Expect(func() {
		_, err = lb.EnsureLoadBalancer(context.Background(), "test", service, nodes)
	}).NotTo(Panic())
	g.Expect(err).To(BeNil())

	g.Expect(fakeCivoClient.LoadBalancers).To(HaveLen(1))
	g.Expect(fakeCivoClient.LoadBalancers[0].Backends).To(HaveLen(1))
	g.Expect(fakeCivoClient.LoadBalancers[0].Backends[0].IP).To(Equal("192.168.1.1"))
}
