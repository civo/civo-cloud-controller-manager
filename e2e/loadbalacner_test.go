package test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestLoadbalacnerBasic(t *testing.T) {

	g := NewGomegaWithT(t)

	mirrorDeploy, err := deployMirrorPods(e2eTest.tenantClient)
	g.Expect(err).ShouldNot(HaveOccurred())

	lbls := map[string]string{"app": "mirror-pod"}
	// Create a service of type: LoadBalacner
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "echo-pods",
			Namespace: "default",
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

func TestLoadbalacnerProxy(t *testing.T) {
	g := NewGomegaWithT(t)

	_, err := deployMirrorPods(e2eTest.tenantClient)
	g.Expect(err).ShouldNot(HaveOccurred())

	lbls := map[string]string{"app": "mirror-pod"}
	// Create a service of type: LoadBalacner
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "echo-pods",
			Namespace: "default",
			Annotations: map[string]string{
				"kubernetes.civo.com/loadbalancer-enable-proxy-protocol": "send-proxy",
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
		return svc.Status.LoadBalancer.Ingress[0].IP
	}, "2m", "5s").ShouldNot(BeEmpty())

	/*
		// Cleanup
		err = cleanUp(mirrorDeploy, svc)
		g.Expect(err).ShouldNot(HaveOccurred())

		g.Eventually(func() error {
			return e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(svc), svc)
		}, "2m", "5s").ShouldNot(BeNil())
	*/
}

func cleanUp(mirrorDeploy *appsv1.Deployment, svc *corev1.Service) error {
	err := e2eTest.tenantClient.Delete(context.TODO(), svc)
	if err != nil {
		return err
	}

	return e2eTest.tenantClient.Delete(context.TODO(), mirrorDeploy)
}

func deployMirrorPods(c client.Client) (*appsv1.Deployment, error) {
	lbls := map[string]string{"app": "mirror-pod"}
	replicas := int32(2)
	mirrorDeploy := &appsv1.Deployment{
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
					Annotations: map[string]string{
						"danm.k8s.io/interfaces": "[{\"tenantNetwork\":\"tenant-vxlan\", \"ip\":\"dynamic\"}]",
					},
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

}
