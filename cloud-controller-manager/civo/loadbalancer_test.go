package civo

import (
	"context"
	"testing"

	"github.com/civo/civogo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	cloudprovider "k8s.io/cloud-provider"
)

func TestGetAlgorithm(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name      string
		algorithm string
		service   *corev1.Service
	}{
		{
			"Algorithm is not set",
			"",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
		},
		{
			"Algorithm should be set to least-connections",
			"least-connections",
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.civo.com/loadbalancer-algorithm": "least-connections",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
		},
		{
			"Algorithm should be set to round-robin",
			"round-robin",
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.civo.com/loadbalancer-algorithm": "round-robin",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alorithm := getAlgorithm(test.service)
			g.Expect(alorithm).To(Equal(test.algorithm))
		})
	}
}

func TestGetFirewallID(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name       string
		firewallID string
		service    *corev1.Service
	}{
		{
			"Firewall is not set",
			"",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
		},
		{
			"Firewall should be set to fc91d382-2609-4f5c-a875-1491776fab8c",
			"fc91d382-2609-4f5c-a875-1491776fab8c",
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.civo.com/firewall-id": "fc91d382-2609-4f5c-a875-1491776fab8c",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			firewall := getFirewallID(test.service)
			g.Expect(firewall).To(Equal(test.firewallID))
		})
	}
}

func TestGetEnableProxyProtocol(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name        string
		enableProxy string
		service     *corev1.Service
	}{
		{
			"EnableProxyProtocol is not set",
			"",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
		},
		{
			"EnableProxyProtocol should be set to send-proxy",
			"send-proxy",
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.civo.com/loadbalancer-enable-proxy-protocol": "send-proxy",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
		},
		{
			"EnableProxyProtocol should be set to send-proxy-v2",
			"send-proxy-v2",
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.civo.com/loadbalancer-enable-proxy-protocol": "send-proxy-v2",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			proxy := getEnableProxyProtocol(test.service)
			g.Expect(proxy).To(Equal(test.enableProxy))
		})
	}
}

func TestUpdateServiceAnnotation(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name       string
		service    *corev1.Service
		annotation string
		value      string
		expected   *corev1.Service
	}{
		{
			"Annotation firewall-id should be set to fc91d382-2609-4f5c-a875-1491776fab8c",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			"kubernetes.civo.com/firewall-id",
			"fc91d382-2609-4f5c-a875-1491776fab8c",
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.civo.com/firewall-id": "fc91d382-2609-4f5c-a875-1491776fab8c",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
		},
		{
			"Annotation for algorithm should be set to round-robin with existing annotation",
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.civo.com/loadbalancer-enable-proxy-protocol": "send-proxy",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			"kubernetes.civo.com/loadbalancer-algorithm",
			"round-robin",
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.civo.com/loadbalancer-enable-proxy-protocol": "send-proxy",
						"kubernetes.civo.com/loadbalancer-algorithm":             "round-robin",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updateServiceAnnotation(test.service, test.annotation, test.value)
			g.Expect(test.service.ObjectMeta.Annotations[test.annotation]).To(Equal(test.expected.ObjectMeta.Annotations[test.annotation]))
			g.Expect(test.service).To(Equal(test.expected))
		})
	}
}

// TestGetLoadBalancerNameFunc tests the getLoadBalancerName function
func TestGetLoadBalancerNameFunc(t *testing.T) {
	g := NewWithT(t)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
	}

	name := getLoadBalancerName("cluster", service)
	g.Expect(name).To(Equal("cluster-test-namespace-test-service"))
}

var _ cloudprovider.LoadBalancer = new(loadbalancer)

// TestGetLoadBalabcer is a test for GetLoadBalancer
func TestGetLoadBalanacer(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name     string
		service  *corev1.Service
		store    []civogo.LoadBalancer
		expected *corev1.LoadBalancerStatus
		exists   bool
		err      error
	}{
		{
			name: "LoadBalancer should be returned",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: corev1.NamespaceDefault,
					Annotations: map[string]string{
						annotationCivoLoadBalancerID:   "6dc1c87d-b8a1-42cd-8fc6-8293378e5715",
						annotationCivoFirewallID:       "fc91d382-2609-4f5c-a875-1491776fab8c",
						annotationCivoClusterID:        "a32fe5eb-1922-43e8-81bc-7f83b4011334",
						annotationCivoLoadBalancerName: "civo-lb-test",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
					Ports: []corev1.ServicePort{
						{
							Name:     "http",
							Protocol: corev1.ProtocolTCP,
							Port:     80,
							NodePort: 30000,
						},
					},
				},
			},
			store: []civogo.LoadBalancer{
				{
					ID:         "6dc1c87d-b8a1-42cd-8fc6-8293378e5715",
					Name:       "civo-lb-test",
					Algorithm:  "round-robin",
					PublicIP:   "192.168.11.11",
					FirewallID: "fc91d382-2609-4f5c-a875-1491776fab8c",
					ClusterID:  "a32fe5eb-1922-43e8-81bc-7f83b4011334",
					State:      statusAvailable,
				},
			},
			expected: &corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{
						IP: "192.168.11.11",
					},
				},
			},
			exists: true,
			err:    nil,
		},
		{
			name: "LoadBalancer should not be returned",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-not-found",
					Namespace: corev1.NamespaceDefault,
					Annotations: map[string]string{
						annotationCivoLoadBalancerID:   "6dc1c87d-b8a1-42cd-8fc6-8293378e5715",
						annotationCivoFirewallID:       "fc91d382-2609-4f5c-a875-1491776fab8c",
						annotationCivoClusterID:        "a32fe5eb-1922-43e8-81bc-7f83b4011334",
						annotationCivoLoadBalancerName: "civo-lb-test-not-found",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
					Ports: []corev1.ServicePort{
						{
							Name:     "http",
							Protocol: corev1.ProtocolTCP,
							Port:     80,
							NodePort: 30000,
						},
					},
				},
			},
			store:    []civogo.LoadBalancer{},
			expected: nil,
			exists:   false,
			err:      nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeCivoClient, err := civogo.NewFakeClient()
			g.Expect(err).To(BeNil())

			fakeCivoClient.LoadBalancers = test.store

			clients := newClients(fakeCivoClient)
			clients.kclient = fake.NewSimpleClientset()

			if _, err := clients.kclient.CoreV1().Services(corev1.NamespaceDefault).Create(context.Background(), test.service, metav1.CreateOptions{}); err != nil {
				t.Fatal(err)
			}

			lb := &loadbalancer{
				client: clients,
			}

			lbStatus, exists, err := lb.GetLoadBalancer(context.TODO(), "test", test.service)
			if test.err == nil {
				g.Expect(err).To(BeNil())
			}

			g.Expect(exists).To(Equal(test.exists))
			g.Expect(lbStatus).To(Equal(test.expected))

			if test.exists {
				svc, err := lb.client.kclient.CoreV1().Services(test.service.Namespace).Get(context.Background(), test.service.Name, metav1.GetOptions{})
				if err != nil {
					t.Fatalf("failed to get service from kube client: %s", err)
				}
				id := svc.Annotations[annotationCivoLoadBalancerID]
				g.Expect(id).To(Equal(test.store[0].ID))
			}
		})
	}
}

// TestGetLoadBalancerName tests the GetLoadBalancerName function
func TestGetLoadBalancerName(t *testing.T) {
	g := NewGomegaWithT(t)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: corev1.NamespaceDefault,
			Annotations: map[string]string{
				annotationCivoLoadBalancerID:   "6dc1c87d-b8a1-42cd-8fc6-8293378e5715",
				annotationCivoFirewallID:       "fc91d382-2609-4f5c-a875-1491776fab8c",
				annotationCivoClusterID:        "a32fe5eb-1922-43e8-81bc-7f83b4011334",
				annotationCivoLoadBalancerName: "civo-lb-test",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					NodePort: 30000,
				},
			},
		},
	}

	lb := &loadbalancer{}

	name := lb.GetLoadBalancerName(context.Background(), "test", service)
	g.Expect(name).To(Equal(service.Annotations[annotationCivoLoadBalancerName]))
}

// TestEnsureLoadBalancer tests the EnsureLoadBalancer function
func TestEnsureLoadBalancer(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name    string
		service *corev1.Service
		nodes   []*corev1.Node
		store   []civogo.LoadBalancer
		cluster []civogo.KubernetesCluster
		setIP   bool
		err     error
	}{
		{
			name: "should create new load balancer",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: corev1.NamespaceDefault,
					Annotations: map[string]string{
						annotationCivoClusterID: "a32fe5eb-1922-43e8-81bc-7f83b4011334",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
					Ports: []corev1.ServicePort{
						{
							Name:     "http",
							Protocol: corev1.ProtocolTCP,
							Port:     80,
							NodePort: 30000,
						},
					},
				},
			},
			nodes: []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Address: "192.168.1.1",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Address: "192.168.1.2",
							},
						},
					},
				},
			},
			cluster: []civogo.KubernetesCluster{
				{
					ID:   "a32fe5eb-1922-43e8-81bc-7f83b4011334",
					Name: "test",
					Instances: []civogo.KubernetesInstance{
						{
							ID:       "11bd4686-5dbf-4e35-b703-75f2864bd6b9",
							Hostname: "node1",
						},
					},
				},
			},
			setIP: true,
			err:   nil,
		},
		{
			name: "should not set ip for proxy protocol",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: corev1.NamespaceDefault,
					Annotations: map[string]string{
						annotationCivoClusterID:                       "a32fe5eb-1922-43e8-81bc-7f83b4011334",
						annotationCivoLoadBalancerEnableProxyProtocol: "send-proxy",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
					Ports: []corev1.ServicePort{
						{
							Name:     "http",
							Protocol: corev1.ProtocolTCP,
							Port:     80,
							NodePort: 30000,
						},
					},
				},
			},
			nodes: []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Address: "192.168.1.1",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Address: "192.168.1.2",
							},
						},
					},
				},
			},
			cluster: []civogo.KubernetesCluster{
				{
					ID:   "a32fe5eb-1922-43e8-81bc-7f83b4011334",
					Name: "test",
					Instances: []civogo.KubernetesInstance{
						{
							ID:       "11bd4686-5dbf-4e35-b703-75f2864bd6b9",
							Hostname: "node1",
						},
					},
				},
			},
			setIP: false,
			err:   nil,
		},
		{
			name: "should update an existing load balancer",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-update",
					Namespace: corev1.NamespaceDefault,
					Annotations: map[string]string{
						annotationCivoLoadBalancerID: "6dc1c87d-b8a1-42cd-8fc6-8293378e5715",
						annotationCivoClusterID:      "a32fe5eb-1922-43e8-81bc-7f83b4011334",
					},
				},
				Spec: corev1.ServiceSpec{
					Type:                  corev1.ServiceTypeLoadBalancer,
					ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
					Ports: []corev1.ServicePort{
						{
							Name:     "http",
							Protocol: corev1.ProtocolTCP,
							Port:     80,
							NodePort: 30000,
						},
					},
				},
			},
			nodes: []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Address: "192.168.1.1",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Address: "192.168.1.2",
							},
						},
					},
				},
			},
			store: []civogo.LoadBalancer{
				{
					ID:                    "6dc1c87d-b8a1-42cd-8fc6-8293378e5715",
					Name:                  "civo-lb-test-update",
					Algorithm:             "round-robin",
					PublicIP:              "192.168.11.11",
					FirewallID:            "fc91d382-2609-4f5c-a875-1491776fab8c",
					ClusterID:             "a32fe5eb-1922-43e8-81bc-7f83b4011334",
					ExternalTrafficPolicy: "Cluster",
					State:                 statusAvailable,
				},
			},
			cluster: []civogo.KubernetesCluster{
				{
					ID:   "a32fe5eb-1922-43e8-81bc-7f83b4011334",
					Name: "test",
					Instances: []civogo.KubernetesInstance{
						{
							ID:       "11bd4686-5dbf-4e35-b703-75f2864bd6b9",
							Hostname: "node1",
						},
					},
				},
			},
			setIP: true,
			err:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeCivoClient, err := civogo.NewFakeClient()
			g.Expect(err).To(BeNil())

			fakeCivoClient.LoadBalancers = test.store
			fakeCivoClient.Clusters = test.cluster

			clients := newClients(fakeCivoClient)
			clients.kclient = fake.NewSimpleClientset()
			ClusterID = test.cluster[0].ID

			if _, err := clients.kclient.CoreV1().Services(corev1.NamespaceDefault).Create(context.Background(), test.service, metav1.CreateOptions{}); err != nil {
				t.Fatal(err)
			}

			lb := &loadbalancer{
				client: clients,
			}

			lbStatus, err := lb.EnsureLoadBalancer(context.Background(), test.cluster[0].Name, test.service, test.nodes)
			g.Expect(err).To(BeNil())

			if test.setIP {
				g.Expect(lbStatus.Ingress[0].IP).NotTo(BeEmpty())
			} else {
				g.Expect(lbStatus.Ingress[0].IP).To(BeEmpty())
			}

			g.Expect(lbStatus.Ingress[0].Hostname).NotTo(BeEmpty())

			svc, err := lb.client.kclient.CoreV1().Services(test.service.Namespace).Get(context.Background(), test.service.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("failed to get service from kube client: %s", err)
			}
			id := svc.Annotations[annotationCivoLoadBalancerID]
			g.Expect(id).NotTo(BeEmpty())

			if svc.Spec.ExternalTrafficPolicy != "" {
				g.Expect(svc.Spec.ExternalTrafficPolicy).To(Equal(corev1.ServiceExternalTrafficPolicyTypeLocal))
			}
		})
	}
}

// TestUpdateLoadBalancer tests the update of a load balancer
func TestUpdateLoadBalancer(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name    string
		service *corev1.Service
		nodes   []*corev1.Node
		store   []civogo.LoadBalancer
		err     error
	}{
		{
			name: "should update an load balancer",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-update",
					Namespace: corev1.NamespaceDefault,
					Annotations: map[string]string{
						annotationCivoLoadBalancerID:                  "6dc1c87d-b8a1-42cd-8fc6-8293378e5715",
						annotationCivoClusterID:                       "a32fe5eb-1922-43e8-81bc-7f83b4011334",
						annotationCivoLoadBalancerAlgorithm:           "least-connections",
						annotationCivoLoadBalancerEnableProxyProtocol: "send-proxy",
					},
				},
				Spec: corev1.ServiceSpec{
					Type:                  corev1.ServiceTypeLoadBalancer,
					ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
					Ports: []corev1.ServicePort{
						{
							Name:     "http",
							Protocol: corev1.ProtocolTCP,
							Port:     80,
							NodePort: 30000,
						},
					},
				},
			},
			nodes: []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Address: "192.168.1.11",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Address: "192.168.1.22",
							},
						},
					},
				},
			},
			store: []civogo.LoadBalancer{
				{
					ID:                    "6dc1c87d-b8a1-42cd-8fc6-8293378e5715",
					Name:                  "civo-lb-test-update",
					Algorithm:             "round-robin",
					PublicIP:              "192.168.11.11",
					FirewallID:            "fc91d382-2609-4f5c-a875-1491776fab8c",
					ClusterID:             "a32fe5eb-1922-43e8-81bc-7f83b4011334",
					ExternalTrafficPolicy: "Cluster",
					State:                 statusAvailable,
					Backends: []civogo.LoadBalancerBackend{
						{
							IP: "192.168.1.1",
						},
						{
							IP: "192.168.1.2",
						},
					},
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeCivoClient, err := civogo.NewFakeClient()
			g.Expect(err).To(BeNil())

			fakeCivoClient.LoadBalancers = test.store

			clients := newClients(fakeCivoClient)
			clients.kclient = fake.NewSimpleClientset()

			if _, err := clients.kclient.CoreV1().Services(corev1.NamespaceDefault).Create(context.Background(), test.service, metav1.CreateOptions{}); err != nil {
				t.Fatal(err)
			}

			lb := &loadbalancer{
				client: clients,
			}

			err = lb.UpdateLoadBalancer(context.Background(), "test", test.service, test.nodes)
			g.Expect(err).To(BeNil())

			svc, err := lb.client.kclient.CoreV1().Services(test.service.Namespace).Get(context.Background(), test.service.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("failed to get service from kube client: %s", err)
			}
			id := svc.Annotations[annotationCivoLoadBalancerID]
			g.Expect(id).NotTo(BeEmpty())

			if svc.Spec.ExternalTrafficPolicy != "" {
				g.Expect(svc.Spec.ExternalTrafficPolicy).To(Equal(corev1.ServiceExternalTrafficPolicyTypeLocal))
			}

			if svc.Annotations[annotationCivoLoadBalancerEnableProxyProtocol] != "" {
				g.Expect(svc.Annotations[annotationCivoLoadBalancerEnableProxyProtocol]).To(Equal("send-proxy"))
			}

			if svc.Annotations[annotationCivoLoadBalancerAlgorithm] != "" {
				g.Expect(svc.Annotations[annotationCivoLoadBalancerAlgorithm]).To(Equal("least-connections"))
			}
		})
	}
}

// TestEnsureLoadBalancerDeleted tests the ensureLoadBalancerDeleted function
func TestEnsureLoadBalancerDeleted(t *testing.T) {
	g := NewGomegaWithT(t)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: corev1.NamespaceDefault,
			Annotations: map[string]string{
				annotationCivoLoadBalancerID:   "6dc1c87d-b8a1-42cd-8fc6-8293378e5715",
				annotationCivoFirewallID:       "fc91d382-2609-4f5c-a875-1491776fab8c",
				annotationCivoClusterID:        "a32fe5eb-1922-43e8-81bc-7f83b4011334",
				annotationCivoLoadBalancerName: "civo-lb-test",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					NodePort: 30000,
				},
			},
		},
	}
	store := []civogo.LoadBalancer{
		{
			ID:                    "6dc1c87d-b8a1-42cd-8fc6-8293378e5715",
			Name:                  "civo-lb-test-update",
			Algorithm:             "round-robin",
			PublicIP:              "192.168.11.11",
			FirewallID:            "fc91d382-2609-4f5c-a875-1491776fab8c",
			ClusterID:             "a32fe5eb-1922-43e8-81bc-7f83b4011334",
			ExternalTrafficPolicy: "Cluster",
			State:                 statusAvailable,
			Backends: []civogo.LoadBalancerBackend{
				{
					IP: "192.168.1.1",
				},
				{
					IP: "192.168.1.2",
				},
			},
		},
	}

	fakeCivoClient, err := civogo.NewFakeClient()
	g.Expect(err).To(BeNil())

	fakeCivoClient.LoadBalancers = store

	clients := newClients(fakeCivoClient)
	clients.kclient = fake.NewSimpleClientset()

	lb := &loadbalancer{
		client: clients,
	}

	err = lb.EnsureLoadBalancerDeleted(context.Background(), "test", service)
	g.Expect(err).To(BeNil())
}

func TestGetProtocol(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := map[string]struct {
		Annotation string
		Result     string
	}{
		"No Annotation Set":        {Annotation: "", Result: "TCP"},
		"Annotation Set":           {Annotation: "HTTP", Result: "HTTP"},
		"Lowercase Annotation Set": {Annotation: "http", Result: "HTTP"},
	}
	svc := &v1.Service{}
	port := v1.ServicePort{Protocol: corev1.ProtocolTCP}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc.Annotations = map[string]string{annotationCivoProtocol: test.Annotation}
			res := getProtocol(svc, port)
			g.Expect(res).Should(Equal(test.Result))
		})

	}
}
