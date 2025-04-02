package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/civo/civogo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TODO: revert the changes before merging.
func aTestLoadbalancerBasic(t *testing.T) {

	g := NewGomegaWithT(t)

	mirrorDeploy, err := deployMirrorPods(e2eTest.tenantClient)
	g.Expect(err).ShouldNot(HaveOccurred())
	defer e2eTest.tenantClient.Delete(context.TODO(), mirrorDeploy)

	lbls := map[string]string{"app": "mirror-pod"}
	// Create a service of type: LoadBalancer
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "echo-pods",
			Namespace: "default",
			Annotations: map[string]string{
				"kubernetes.civo.com/firewall-id": e2eTest.cluster.FirewallID,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Protocol: "TCP", Port: 80, TargetPort: intstr.FromInt(8080)},
				{Name: "https", Protocol: "TCP", Port: 443, TargetPort: intstr.FromInt(8443)},
			},
			Selector: lbls,
			Type:     "LoadBalancer",
		},
	}

	fmt.Println("Creating Service")
	err = e2eTest.tenantClient.Create(context.TODO(), svc)
	g.Expect(err).ShouldNot(HaveOccurred())
	defer e2eTest.tenantClient.Delete(context.TODO(), svc)

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(svc), svc)
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return ""
		}
		return svc.Status.LoadBalancer.Ingress[0].IP
	}, "2m", "5s").ShouldNot(BeEmpty())

	// Cleanup
	err = cleanUp(mirrorDeploy, svc)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() error {
		return e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(svc), svc)
	}, "2m", "5s").ShouldNot(BeNil())
}

// TODO: revert the changes before merging.
func aTestLoadbalancerProxy(t *testing.T) {
	g := NewGomegaWithT(t)

	mirrorDeploy, err := deployMirrorPods(e2eTest.tenantClient)
	g.Expect(err).ShouldNot(HaveOccurred())

	lbls := map[string]string{"app": "mirror-pod"}
	// Create a service of type: LoadBalancer
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "echo-pods",
			Namespace: "default",
			Annotations: map[string]string{
				"kubernetes.civo.com/loadbalancer-enable-proxy-protocol": "send-proxy",
				"kubernetes.civo.com/firewall-id":                        e2eTest.cluster.FirewallID,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Protocol: "TCP", Port: 80, TargetPort: intstr.FromInt(8081)},
				{Name: "https", Protocol: "TCP", Port: 443, TargetPort: intstr.FromInt(8444)},
			},
			Selector: lbls,
			Type:     "LoadBalancer",
		},
	}

	fmt.Println("Creating Service")
	err = e2eTest.tenantClient.Create(context.TODO(), svc)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(svc), svc)
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return ""
		}
		return svc.Status.LoadBalancer.Ingress[0].Hostname
	}, "5m", "5s").ShouldNot(BeEmpty())

	// Cleanup
	err = cleanUp(mirrorDeploy, svc)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() error {
		return e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(svc), svc)
	}, "2m", "5s").ShouldNot(BeNil())
}

// TODO: revert the changes before merging.
func aTestLoadbalancerHTTPForwardFor(t *testing.T) {
	g := NewGomegaWithT(t)

	mirrorDeploy, err := deployMirrorPods(e2eTest.tenantClient)
	g.Expect(err).ShouldNot(HaveOccurred())
	defer e2eTest.tenantClient.Delete(context.TODO(), mirrorDeploy)

	lbls := map[string]string{"app": "mirror-pod"}
	// Create a service of type: LoadBalancer
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "echo-pods",
			Namespace: "default",
			Annotations: map[string]string{
				"kubernetes.civo.com/loadbalancer-enable-proxy-protocol": "send-proxy",
				"kubernetes.civo.com/firewall-id":                        e2eTest.cluster.FirewallID,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Protocol: "TCP", Port: 80, TargetPort: intstr.FromInt(8080)},
			},
			Selector:              lbls,
			Type:                  "LoadBalancer",
			ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
		},
	}

	fmt.Println("Creating Service")
	err = e2eTest.tenantClient.Create(context.TODO(), svc)
	g.Expect(err).ShouldNot(HaveOccurred())
	defer e2eTest.tenantClient.Delete(context.TODO(), svc)

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(svc), svc)
		g.Expect(err).ShouldNot(HaveOccurred())
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return ""
		}
		return svc.Status.LoadBalancer.Ingress[0].IP
	}, "5m", "5s").ShouldNot(BeEmpty())

	// Make a Web Request
	err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(svc), svc)
	g.Expect(err).ShouldNot(HaveOccurred())

	resp, err := http.Get("http://" + svc.Status.LoadBalancer.Ingress[0].IP)
	g.Expect(err).ShouldNot(HaveOccurred())
	body, err := ioutil.ReadAll(resp.Body)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(string(body)).Should(ContainSubstring("x-forwarded-for"))

	// Cleanup
	err = cleanUp(mirrorDeploy, svc)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() error {
		return e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(svc), svc)
	}, "2m", "5s").ShouldNot(BeNil())
}

func TestLoadbalancerReservedIP(t *testing.T) {
	g := NewGomegaWithT(t)

	_, err := deployMirrorPods(e2eTest.tenantClient)
	g.Expect(err).ShouldNot(HaveOccurred())

	fmt.Println("Create a reserved IP for e2e test (if it doesn't exist)")
	ip, err := getOrCreateIP(e2eTest.civo)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		ip, err = e2eTest.civo.GetIP(ip.ID)
		return ip.IP
	}, "2m", "5s").ShouldNot(BeEmpty())

	fmt.Println("Creating Service")
	svc, err := getOrCreateSvc(e2eTest.tenantClient)
	g.Expect(err).ShouldNot(HaveOccurred())

	patchSvc := &corev1.Service{}
	err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(svc), patchSvc)
	originalSvc := svc.DeepCopy()
	if patchSvc.Annotations == nil {
		patchSvc.Annotations = make(map[string]string, 0)
	}
	patchSvc.Annotations = map[string]string{
		"kubernetes.civo.com/ipv4-address": ip.IP,
	}

	fmt.Println("Updating service with reserved IP annotation")
	err = e2eTest.tenantClient.Patch(context.TODO(), patchSvc, client.MergeFrom(originalSvc))
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(svc), svc)
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return ""
		}
		return svc.Status.LoadBalancer.Ingress[0].IP
	}, "5m", "5s").Should(Equal(ip.IP))

	// Unassign reserved IP
	fmt.Println("Unassigning IP from LB")
	svc.Annotations = nil
	err = e2eTest.tenantClient.Update(context.TODO(), svc)
	g.Expect(err).ShouldNot(HaveOccurred())

	fmt.Println("Waiting for auto-assigned IP to be attached to LB")
	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(svc), svc)
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return ""
		}
		return svc.Status.LoadBalancer.Ingress[0].IP
	}, "5m", "5s").ShouldNot(Equal(ip.IP))

	// To make sure an auto-assigned IP is actually assigned to the LB
	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(svc), svc)
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return ""
		}
		return svc.Status.LoadBalancer.Ingress[0].IP
	}, "5m", "5s").ShouldNot(BeEmpty())

}

func cleanUp(mirrorDeploy *appsv1.Deployment, svc *corev1.Service) error {
	e2eTest.tenantClient.Delete(context.TODO(), svc)

	return e2eTest.tenantClient.Delete(context.TODO(), mirrorDeploy)
}

func getOrCreateIP(c *civogo.Client) (*civogo.IP, error) {
	ip, err := c.FindIP("ccm-e2e-test-ip")
	if err != nil && civogo.ZeroMatchesError.Is(err) {
		ip, err = c.NewIP(&civogo.CreateIPRequest{
			Name: "ccm-e2e-test-ip",
		})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return ip, err
}

func getOrCreateSvc(c client.Client) (*corev1.Service, error) {
	svc := &corev1.Service{}
	err := c.Get(context.TODO(), client.ObjectKey{Name: "echo-pods", Namespace: "default"}, svc)
	if err != nil && errors.IsNotFound(err) {
		lbls := map[string]string{"app": "mirror-pod"}
		// Create a service of type: LoadBalancer
		svc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "echo-pods",
				Namespace: "default",
				Annotations: map[string]string{
					"kubernetes.civo.com/firewall-id": e2eTest.cluster.FirewallID,
				},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{Name: "http", Protocol: "TCP", Port: 80, TargetPort: intstr.FromInt(8081)},
					{Name: "https", Protocol: "TCP", Port: 443, TargetPort: intstr.FromInt(8444)},
				},
				Selector: lbls,
				Type:     "LoadBalancer",
			},
		}
		err = c.Create(context.TODO(), svc)
		return svc, err
	} else if err != nil {
		return nil, err
	}
	return nil, err
}

func deployMirrorPods(c client.Client) (*appsv1.Deployment, error) {
	mirrorDeploy := &appsv1.Deployment{}
	err := c.Get(context.TODO(), client.ObjectKey{Name: "echo-pods", Namespace: "default"}, mirrorDeploy)
	if err != nil && errors.IsNotFound(err) {
		lbls := map[string]string{"app": "mirror-pod"}
		replicas := int32(2)
		mirrorDeploy = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "echo-pods",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: lbls,
				},
				Replicas: &replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: lbls,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:            "mirror-pod",
								Image:           "dmajrekar/nginx-echo:latest",
								ImagePullPolicy: corev1.PullIfNotPresent,
								Ports: []corev1.ContainerPort{
									{Protocol: "TCP", ContainerPort: 8080},
									{Protocol: "TCP", ContainerPort: 8081},
									{Protocol: "TCP", ContainerPort: 8443},
									{Protocol: "TCP", ContainerPort: 8444},
								},
							},
						},
					},
				},
			},
		}

		fmt.Println("Creating mirror deployment")
		err := c.Create(context.TODO(), mirrorDeploy)
		return mirrorDeploy, err
	} else if err != nil {
		return nil, err
	}

	return mirrorDeploy, err
}
