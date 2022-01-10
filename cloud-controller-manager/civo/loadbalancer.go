package civo

import (
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"

	"github.com/civo/civogo"
)

type loadbalancer struct {
	civoClient *civogo.Client
}

// TODO: Break this up into different interfaces (LB, etc) when we have more than one type of service
// GetLoadBalancer returns whether the specified load balancer exists, and
// if so, what its status is.
// Implementations must treat the *v1.Service parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (l *loadbalancer) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (*v1.LoadBalancerStatus, bool, error) {
	civolb, err := l.civoClient.FindLoadBalancer(getLoadBalancerName(clusterName, service))
	if err != nil && !strings.Contains(err.Error(), string(civogo.ZeroMatchesError)) {
		return nil, false, err
	}

	if err != nil && strings.Contains(err.Error(), string(civogo.ZeroMatchesError)) {
		return nil, false, nil
	}

	return &v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{{IP: civolb.PublicIP}}}, true, nil
}

// GetLoadBalancerName returns the name of the load balancer. Implementations must treat the
// *v1.Service parameter as read-only and not modify it.
func (*loadbalancer) GetLoadBalancerName(_ context.Context, clusterName string, service *v1.Service) string {
	return getLoadBalancerName(clusterName, service)
}

// EnsureLoadBalancer creates a new load balancer 'name', or updates the existing one. Returns the status of the balancer
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (l *loadbalancer) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	civolb, err := l.civoClient.FindLoadBalancer(getLoadBalancerName(clusterName, service))
	if err != nil && !strings.Contains(err.Error(), string(civogo.ZeroMatchesError)) {
		return nil, err
	}

	// CivLB has been found
	if err == nil {
		lbc := civogo.LoadBalancerConfig{
			ExternalTrafficPolicy: string(service.Spec.ExternalTrafficPolicy),
		}
		updatedlb, err := l.civoClient.UpdateLoadBalancer(civolb.ID, &lbc)
		if err != nil {
			return nil, err
		}

		return &v1.LoadBalancerStatus{
			Ingress: []v1.LoadBalancerIngress{
				{IP: updatedlb.PublicIP},
			},
		}, nil
	}

	err = createLoadBalancer(ctx, clusterName, service, nodes, l.civoClient)
	if err != nil {
		return nil, err
	}

	civolb, err = l.civoClient.FindLoadBalancer(getLoadBalancerName(clusterName, service))
	if err != nil && !strings.Contains(err.Error(), string(civogo.ZeroMatchesError)) {
		return nil, err
	}

	if civolb.State != "available" {
		return nil, nil
	}
	return &v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{
			{IP: civolb.PublicIP},
		},
	}, nil
}

// UpdateLoadBalancer updates hosts under the specified load balancer.
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (l *loadbalancer) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	// TODO: Do this better ...
	// Delete the LoadBalancer
	err := l.EnsureLoadBalancerDeleted(ctx, clusterName, service)
	if err != nil {
		return err
	}

	return createLoadBalancer(ctx, clusterName, service, nodes, l.civoClient)
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
	civolb, err := l.civoClient.FindLoadBalancer(getLoadBalancerName(clusterName, service))
	if err != nil && !strings.Contains(err.Error(), string(civogo.ZeroMatchesError)) {
		return err
	}

	_, err = l.civoClient.DeleteLoadBalancer(civolb.ID)
	if err != nil {
		return err
	}

	return nil
}

func getLoadBalancerName(clusterName string, service *v1.Service) string {
	return fmt.Sprintf("%s-%s-%s", clusterName, service.Namespace, service.Name)
}

func createLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node, civoClient *civogo.Client) error {
	lbc := civogo.LoadBalancerConfig{
		Name:                  getLoadBalancerName(clusterName, service),
		ExternalTrafficPolicy: string(service.Spec.ExternalTrafficPolicy),
		Region:                Region,
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

	if _, err := civoClient.CreateLoadBalancer(&lbc); err != nil {
		return err
	}

	return nil
}
