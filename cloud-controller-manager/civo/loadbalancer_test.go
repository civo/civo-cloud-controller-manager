package civo

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
