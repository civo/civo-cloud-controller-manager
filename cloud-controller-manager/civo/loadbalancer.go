package civo

import (
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"

	"github.com/civo/civogo"
)

const (
	// status for Civo load balancer
	statusAvailable = "available"

	// annotationCivoClusterID is the annotation specifying the CivoCluster ID.
	annotationCivoClusterID = "kubernetes.civo.com/cluster-id"

	// annotationCivoFirewallID is the annotation specifying the CivoFirewall ID.
	annotationCivoFirewallID = "kubernetes.civo.com/firewall-id"

	// annotationCivoLoadBalancerEnableProxyProtocol is the annotation specifying whether PROXY protocol should be enabled.
	annotationCivoLoadBalancerEnableProxyProtocol = "kubernetes.civo.com/loadbalancer-enable-proxy-protocol"

	// annotationCivoLoadBalancerID is the annotation specifying the CivoLoadbalancer ID.
	annotationCivoLoadBalancerID = "kubernetes.civo.com/loadbalancer-id"

	// annotationCivoLoadBalancerName is the annotation specifying the CivoLoadbalancer name.
	annotationCivoLoadBalancerName = "kubernetes.civo.com/loadbalancer-name"

	// annotationCivoLoadBalancerAlgorithm is the annotation specifying the CivoLoadbalancer algorith.
	annotationCivoLoadBalancerAlgorithm = "kubernetes.civo.com/loadbalancer-algorithm"
)

type loadbalancer struct {
	client *clients
}

// newLoadbalancers returns a cloudprovider.LoadBalancer whose concrete type is a *loadbalancer.
func newLoadBalancers(c *clients) cloudprovider.LoadBalancer {
	return &loadbalancer{
		client: c,
	}
}

// TODO: Break this up into different interfaces (LB, etc) when we have more than one type of service
// GetLoadBalancer returns whether the specified load balancer exists, and
// if so, what its status is.
// Implementations must treat the *v1.Service parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (l *loadbalancer) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (*v1.LoadBalancerStatus, bool, error) {
	civolb, err := getLoadBalancer(ctx, l.client.civoClient, l.client.kclient, clusterName, service)
	if err != nil {
		if strings.Contains(err.Error(), string(civogo.ZeroMatchesError)) || strings.Contains(err.Error(), string(civogo.DatabaseLoadBalancerNotFoundError)) {
			return nil, false, nil
		}
		klog.Errorf("Unable to get loadbalancer, error: %v", err)
		return nil, false, err
	}

	return &v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{
			{
				IP: civolb.PublicIP,
			},
		},
	}, true, nil
}

// GetLoadBalancerName returns the name of the load balancer. Implementations must treat the
// *v1.Service parameter as read-only and not modify it.
func (*loadbalancer) GetLoadBalancerName(_ context.Context, clusterName string, service *v1.Service) string {
	return service.Annotations[annotationCivoLoadBalancerName]
}

// EnsureLoadBalancer creates a new load balancer 'name', or updates the existing one. Returns the status of the balancer
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (l *loadbalancer) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	civolb, err := getLoadBalancer(ctx, l.client.civoClient, l.client.kclient, clusterName, service)
	if err != nil && !strings.Contains(err.Error(), string(civogo.ZeroMatchesError)) && !strings.Contains(err.Error(), string(civogo.DatabaseLoadBalancerNotFoundError)) {
		klog.Errorf("Unable to create loadbalancer, error: %v", err)
		return nil, err
	}

	// CivLB has been found
	if err == nil {
		lbuc := civogo.LoadBalancerUpdateConfig{
			ExternalTrafficPolicy: string(service.Spec.ExternalTrafficPolicy),
			Region:                Region,
		}

		if enableProxyProtocol := getEnableProxyProtocol(service); enableProxyProtocol != "" {
			lbuc.EnableProxyProtocol = enableProxyProtocol
		}
		if algorithm := getAlgorithm(service); algorithm != "" {
			lbuc.Algorithm = algorithm
		}
		if firewallID := getFirewallID(service); firewallID != "" {
			lbuc.FirewallID = firewallID
		}

		updatedlb, err := l.client.civoClient.UpdateLoadBalancer(civolb.ID, &lbuc)
		if err != nil {
			klog.Errorf("Unable to update loadbalancer, error: %v", err)
			return nil, err
		}

		return &v1.LoadBalancerStatus{
			Ingress: []v1.LoadBalancerIngress{
				{
					IP: updatedlb.PublicIP,
				},
			},
		}, nil
	}

	err = createLoadBalancer(ctx, clusterName, service, nodes, l.client.civoClient, l.client.kclient)
	if err != nil {
		return nil, err
	}

	civolb, err = getLoadBalancer(ctx, l.client.civoClient, l.client.kclient, clusterName, service)
	if err != nil && !strings.Contains(err.Error(), string(civogo.ZeroMatchesError)) {
		klog.Errorf("Unable to get loadbalancer, error: %v", err)
		return nil, err
	}

	if civolb.State != statusAvailable {
		klog.Errorf("Loadbalancer is not available, state: %s", civolb.State)
		return nil, fmt.Errorf("loadbalancer is not yet available, current state: %s", civolb.State)
	}

	return &v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{
			{
				IP: civolb.PublicIP,
			},
		},
	}, nil
}

// UpdateLoadBalancer updates hosts under the specified load balancer.
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (l *loadbalancer) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	civolb, err := getLoadBalancer(ctx, l.client.civoClient, l.client.kclient, clusterName, service)
	if err != nil {
		if strings.Contains(err.Error(), string(civogo.ZeroMatchesError)) || strings.Contains(err.Error(), string(civogo.DatabaseLoadBalancerNotFoundError)) {
			return nil
		}
		klog.Errorf("Unable to get loadbalancer, error: %v", err)
		return err
	}

	lbuc := civogo.LoadBalancerUpdateConfig{
		ExternalTrafficPolicy: string(service.Spec.ExternalTrafficPolicy),
		Region:                Region,
	}

	if enableProxyProtocol := getEnableProxyProtocol(service); enableProxyProtocol != "" {
		lbuc.EnableProxyProtocol = enableProxyProtocol
	}
	if algorithm := getAlgorithm(service); algorithm != "" {
		lbuc.Algorithm = algorithm
	}
	if firewallID := getFirewallID(service); firewallID != "" {
		lbuc.FirewallID = firewallID
	}

	backends := []civogo.LoadBalancerBackendConfig{}
	for _, port := range service.Spec.Ports {
		for _, node := range nodes {
			backends = append(backends, civogo.LoadBalancerBackendConfig{
				IP:              node.Status.Addresses[0].Address,
				Protocol:        string(port.Protocol),
				SourcePort:      port.Port,
				TargetPort:      port.NodePort,
				HealthCheckPort: service.Spec.HealthCheckNodePort,
			})
		}
	}
	lbuc.Backends = backends

	ulb, err := l.client.civoClient.UpdateLoadBalancer(civolb.ID, &lbuc)
	if err != nil {
		klog.Errorf("Unable to update loadbalancer, error: %v", err)
		return err
	}

	patcher := newServicePatcher(l.client.kclient, service)
	defer func() { err = patcher.Patch(ctx, err) }()

	if civolb.Algorithm != ulb.Algorithm {
		updateServiceAnnotation(service, annotationCivoLoadBalancerAlgorithm, ulb.Algorithm)
	}
	if civolb.FirewallID != ulb.FirewallID {
		updateServiceAnnotation(service, annotationCivoFirewallID, ulb.FirewallID)
	}
	if civolb.EnableProxyProtocol != ulb.EnableProxyProtocol {
		updateServiceAnnotation(service, annotationCivoLoadBalancerEnableProxyProtocol, ulb.EnableProxyProtocol)
	}

	return nil
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it
// exists, returning nil if the load balancer specified either didn't exist or
// was successfully deleted.
// This construction is useful because many cloud providers' load balancers
// have multiple underlying components, meaning a Get could say that the LB
// doesn't exist even if some part of it is still laying around.
// Implementations must treat the *v1.Service parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (l *loadbalancer) EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error {
	civolb, err := getLoadBalancer(ctx, l.client.civoClient, l.client.kclient, clusterName, service)
	if err != nil {
		if strings.Contains(err.Error(), string(civogo.ZeroMatchesError)) || strings.Contains(err.Error(), string(civogo.DatabaseLoadBalancerNotFoundError)) {
			return nil
		}
		klog.Errorf("Unable to get loadbalancer, error: %v", err)
		return err
	}

	_, err = l.client.civoClient.DeleteLoadBalancer(civolb.ID)
	if err != nil {
		klog.Errorf("Unable to delete loadbalancer, error: %v", err)
		return err
	}

	return nil
}

func getLoadBalancerName(clusterName string, service *v1.Service) string {
	return fmt.Sprintf("%s-%s-%s", clusterName, service.Namespace, service.Name)
}

func getLoadBalancer(ctx context.Context, c *civogo.Client, kclient kubernetes.Interface, clusterName string, service *v1.Service) (*civogo.LoadBalancer, error) {
	var err error
	var civolb *civogo.LoadBalancer
	if id, ok := service.Annotations[annotationCivoLoadBalancerID]; ok {
		civolb, err = c.GetLoadBalancer(id)
	} else if name, ok := service.Annotations[annotationCivoLoadBalancerName]; ok {
		civolb, err = c.GetLoadBalancer(name)
	} else {
		lbName := getLoadBalancerName(clusterName, service)
		civolb, err = c.FindLoadBalancer(lbName)
		if err == nil {
			patcher := newServicePatcher(kclient, service)
			defer func() { err = patcher.Patch(ctx, err) }()

			updateServiceAnnotation(service, annotationCivoLoadBalancerID, civolb.ID)
			updateServiceAnnotation(service, annotationCivoLoadBalancerName, civolb.Name)
		}
	}

	return civolb, err
}

func createLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node, civoClient *civogo.Client, kclient kubernetes.Interface) error {
	lbc := civogo.LoadBalancerConfig{
		Name:                  getLoadBalancerName(clusterName, service),
		ClusterID:             ClusterID,
		ExternalTrafficPolicy: string(service.Spec.ExternalTrafficPolicy),
		Region:                Region,
	}

	if enableProxyProtocol := getEnableProxyProtocol(service); enableProxyProtocol != "" {
		lbc.EnableProxyProtocol = enableProxyProtocol
	}
	if algorithm := getAlgorithm(service); algorithm != "" {
		lbc.Algorithm = algorithm
	}
	if firewallID := getFirewallID(service); firewallID != "" {
		lbc.FirewallID = firewallID
	}

	backends := []civogo.LoadBalancerBackendConfig{}
	for _, port := range service.Spec.Ports {
		for _, node := range nodes {
			backends = append(backends, civogo.LoadBalancerBackendConfig{
				IP:              node.Status.Addresses[0].Address,
				Protocol:        string(port.Protocol),
				SourcePort:      port.Port,
				TargetPort:      port.NodePort,
				HealthCheckPort: service.Spec.HealthCheckNodePort,
			})
		}
	}
	lbc.Backends = backends

	lb, err := civoClient.CreateLoadBalancer(&lbc)
	if err != nil {
		return err
	}

	patcher := newServicePatcher(kclient, service)
	defer func() { err = patcher.Patch(ctx, err) }()

	updateServiceAnnotation(service, annotationCivoClusterID, ClusterID)
	updateServiceAnnotation(service, annotationCivoFirewallID, lb.FirewallID)
	updateServiceAnnotation(service, annotationCivoLoadBalancerID, lb.ID)
	updateServiceAnnotation(service, annotationCivoLoadBalancerName, lb.Name)
	updateServiceAnnotation(service, annotationCivoLoadBalancerAlgorithm, lb.Algorithm)

	if lb.EnableProxyProtocol != "" {
		updateServiceAnnotation(service, annotationCivoLoadBalancerEnableProxyProtocol, lb.EnableProxyProtocol)
	}

	return nil
}

func updateServiceAnnotation(service *v1.Service, key, value string) {
	if service.ObjectMeta.Annotations == nil {
		service.ObjectMeta.Annotations = make(map[string]string)
	}
	service.ObjectMeta.Annotations[key] = value
}

// getEnableProxyProtocol returns the enableProxyProtocol value from the service annotation.
func getEnableProxyProtocol(service *v1.Service) string {
	epp, ok := service.Annotations[annotationCivoLoadBalancerEnableProxyProtocol]
	if !ok {
		return ""
	}

	return epp
}

// getAlgorithm returns the algorithm value from the service annotation.
func getAlgorithm(service *v1.Service) string {
	algorithm, ok := service.Annotations[annotationCivoLoadBalancerAlgorithm]
	if !ok {
		return ""
	}

	return algorithm
}

// getFirewallID returns the firewallID value from the service annotation.
func getFirewallID(service *v1.Service) string {
	firewallID, ok := service.Annotations[annotationCivoFirewallID]
	if !ok {
		return ""
	}

	return firewallID
}
