package test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/civo/civogo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestLoadbalancerBasic(t *testing.T) {
	ctx := t.Context()
	g := NewGomegaWithT(t)

	mirrorDeploy, err := deployMirrorPods(ctx, e2eTest.tenantClient)
	g.Expect(err).ShouldNot(HaveOccurred())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		e2eTest.tenantClient.Delete(ctx, mirrorDeploy)
	})

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
	err = e2eTest.tenantClient.Create(ctx, svc)
	g.Expect(err).ShouldNot(HaveOccurred())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		e2eTest.tenantClient.Delete(ctx, svc)
	})

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(ctx, client.ObjectKeyFromObject(svc), svc)
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return ""
		}
		return svc.Status.LoadBalancer.Ingress[0].IP
	}, "2m", "5s").ShouldNot(BeEmpty())

	// Cleanup
	err = cleanUp(ctx, mirrorDeploy, svc)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() error {
		return e2eTest.tenantClient.Get(ctx, client.ObjectKeyFromObject(svc), svc)
	}, "2m", "5s").ShouldNot(BeNil())
}

func TestLoadbalancerProxy(t *testing.T) {
	ctx := t.Context()
	g := NewGomegaWithT(t)

	mirrorDeploy, err := deployMirrorPods(ctx, e2eTest.tenantClient)
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
	err = e2eTest.tenantClient.Create(ctx, svc)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(ctx, client.ObjectKeyFromObject(svc), svc)
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return ""
		}
		return svc.Status.LoadBalancer.Ingress[0].Hostname
	}, "5m", "5s").ShouldNot(BeEmpty())

	// Cleanup
	err = cleanUp(ctx, mirrorDeploy, svc)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() error {
		return e2eTest.tenantClient.Get(ctx, client.ObjectKeyFromObject(svc), svc)
	}, "2m", "5s").ShouldNot(BeNil())
}

func TestLoadbalancerEnableProxyProtocol(t *testing.T) {
	ctx := t.Context()
	g := NewGomegaWithT(t)

	mirrorDeploy, err := deployMirrorPods(ctx, e2eTest.tenantClient)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		err := cleanUp(ctx, mirrorDeploy, nil)
		g.Expect(err).ShouldNot(HaveOccurred())
	})
	g.Expect(err).ShouldNot(HaveOccurred())

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
			},
			Selector:              mirrorDeploy.Spec.Template.Labels,
			Type:                  "LoadBalancer",
			ExternalTrafficPolicy: "Local",
		},
	}

	t.Log("Creating Service")
	err = e2eTest.tenantClient.Create(ctx, svc)
	g.Expect(err).ShouldNot(HaveOccurred())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		err := cleanUp(ctx, nil, svc)
		g.Expect(err).ShouldNot(HaveOccurred())

		// Service deletion takes time, so make sure to check until it is fully deleted just in case.
		g.Eventually(func() bool {
			err := e2eTest.tenantClient.Get(ctx, client.ObjectKeyFromObject(svc), svc)
			return apierrors.IsNotFound(err)
		}, "2m", "5s").Should(BeTrue())
	})

	g.Eventually(func() string {
		_ = e2eTest.tenantClient.Get(ctx, client.ObjectKeyFromObject(svc), svc)
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return ""
		}
		return svc.Status.LoadBalancer.Ingress[0].Hostname
	}, "5m", "2m", "5s").ShouldNot(BeEmpty())

	resp, err := http.Get("http://" + svc.Status.LoadBalancer.Ingress[0].Hostname)
	g.Expect(err).ShouldNot(HaveOccurred())
	t.Cleanup(func() {
		if resp != nil {
			g.Expect(resp.Body.Close()).ShouldNot(HaveOccurred())
		}
	})
	b, err := io.ReadAll(resp.Body)
	g.Expect(err).ShouldNot(HaveOccurred())

	// NOTE: https://github.com/DMajrekar/dockerfiles/blob/main/nginx-echo/nginx.conf#L45-L48
	g.Expect(string(b)).Should(ContainSubstring("proxy_client_address"))
}

func TestLoadbalancerReservedIP(t *testing.T) {
	ctx := t.Context()
	g := NewGomegaWithT(t)

	mirrorDeploy, err := deployMirrorPods(ctx, e2eTest.tenantClient)
	g.Expect(err).ShouldNot(HaveOccurred())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		err := cleanUp(ctx, mirrorDeploy, nil)
		g.Expect(err).ShouldNot(HaveOccurred())
	})

	fmt.Println("Create a reserved IP for e2e test (if it doesn't exist)")
	ip, ipCleanup, err := getOrCreateIP(e2eTest.civo)
	g.Expect(err).ShouldNot(HaveOccurred())
	if ipCleanup != nil {
		t.Cleanup(func() {
			err = ipCleanup()
			g.Expect(err).ShouldNot(HaveOccurred())
		})
	}

	g.Eventually(func() string {
		ip, err = e2eTest.civo.GetIP(ip.ID)
		return ip.IP
	}, "2m", "5s").ShouldNot(BeEmpty())

	fmt.Println("Creating Service")
	svc, svcCreated, err := getOrCreateSvc(ctx, e2eTest.tenantClient)
	g.Expect(err).ShouldNot(HaveOccurred())
	if svcCreated {
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			err := cleanUp(ctx, nil, svc)
			g.Expect(err).ShouldNot(HaveOccurred())

			// Service deletion takes time, so make sure to check until it is fully deleted just in case.
			g.Eventually(func() bool {
				err := e2eTest.tenantClient.Get(ctx, client.ObjectKeyFromObject(svc), svc)
				return apierrors.IsNotFound(err)
			}, "2m", "5s").Should(BeTrue())
		})
	}

	patchSvc := &corev1.Service{}
	err = e2eTest.tenantClient.Get(ctx, client.ObjectKeyFromObject(svc), patchSvc)
	originalSvc := svc.DeepCopy()
	if patchSvc.Annotations == nil {
		patchSvc.Annotations = make(map[string]string, 0)
	}
	patchSvc.Annotations = map[string]string{
		"kubernetes.civo.com/ipv4-address": ip.IP,
	}

	fmt.Println("Updating service with reserved IP annotation")
	err = e2eTest.tenantClient.Patch(ctx, patchSvc, client.MergeFrom(originalSvc))
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(ctx, client.ObjectKeyFromObject(svc), svc)
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return ""
		}
		return svc.Status.LoadBalancer.Ingress[0].IP
	}, "5m", "5s").Should(Equal(ip.IP))

	// Unassign reserved IP
	fmt.Println("Unassigning IP from LB")
	svc.Annotations = nil
	err = e2eTest.tenantClient.Update(ctx, svc)
	g.Expect(err).ShouldNot(HaveOccurred())

	fmt.Println("Waiting for auto-assigned IP to be attached to LB")
	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(ctx, client.ObjectKeyFromObject(svc), svc)
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return ""
		}
		return svc.Status.LoadBalancer.Ingress[0].IP
	}, "5m", "5s").ShouldNot(Equal(ip.IP))

	// To make sure an auto-assigned IP is actually assigned to the LB
	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(ctx, client.ObjectKeyFromObject(svc), svc)
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return ""
		}
		return svc.Status.LoadBalancer.Ingress[0].IP
	}, "5m", "5s").ShouldNot(BeEmpty())

}

func cleanUp(ctx context.Context, mirrorDeploy *appsv1.Deployment, svc *corev1.Service) (err error) {
	if svc != nil {
		err = e2eTest.tenantClient.Delete(ctx, svc)
	}
	if mirrorDeploy != nil {
		err = errors.Join(err, e2eTest.tenantClient.Delete(ctx, mirrorDeploy))
	}
	return err
}

func getOrCreateIP(c *civogo.Client) (_ *civogo.IP, cleanup func() error, _ error) {
	ip, err := c.FindIP("ccm-e2e-test-ip")
	if err != nil && civogo.ZeroMatchesError.Is(err) {
		ip, err = c.NewIP(&civogo.CreateIPRequest{
			Name:   "ccm-e2e-test-ip",
			Region: e2eTest.civo.Region,
		})
		if err != nil {
			return nil, nil, err
		}
		cleanup = func() error {
			_, err := c.DeleteIP(ip.ID)
			return err
		}
	} else if err != nil {
		return nil, nil, err
	}
	return ip, cleanup, err
}

func getOrCreateSvc(ctx context.Context, c client.Client) (svc *corev1.Service, created bool, _ error) {
	svc = &corev1.Service{}
	err := c.Get(ctx, client.ObjectKey{Name: "echo-pods", Namespace: "default"}, svc)
	if err != nil && apierrors.IsNotFound(err) {
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
		if err := c.Create(ctx, svc); err != nil {
			return nil, false, err
		}
		return svc, true, nil
	} else if err != nil {
		return nil, false, err
	}
	return svc, false, err
}

func deployMirrorPods(ctx context.Context, c client.Client) (*appsv1.Deployment, error) {
	mirrorDeploy := &appsv1.Deployment{}
	err := c.Get(ctx, client.ObjectKey{Name: "echo-pods", Namespace: "default"}, mirrorDeploy)
	if err != nil && apierrors.IsNotFound(err) {
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
		err := c.Create(ctx, mirrorDeploy)
		return mirrorDeploy, err
	} else if err != nil {
		return nil, err
	}

	return mirrorDeploy, err
}
