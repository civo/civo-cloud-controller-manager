package civo

import (
	"context"

	"github.com/civo/civo-cloud-controller-manager/cloud-controller-manager/pkg/utils"
	"github.com/civo/civogo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
)

const (
	// status for civo instance
	statusActive = "ACTIVE"
)

type instances struct {
	client *clients
}

func newInstances(c *clients) cloudprovider.Instances {
	return &instances{
		client: c,
	}
}

// NodeAddresses returns the addresses of the specified instance.
func (i *instances) NodeAddresses(ctx context.Context, name types.NodeName) ([]v1.NodeAddress, error) {
	instance, err := utils.CivoInstanceFromName(ClusterID, string(name), i.client.civoClient)
	if err != nil {
		klog.Errorf("Unable to get instance by name: %s, error: %v", string(name), err)
		return nil, err
	}

	// Return Public and Private Addresses
	return getAddressessFromCivoInstance(instance), nil

}

func getAddressessFromCivoInstance(instance civogo.Instance) []v1.NodeAddress {
	nodeAdresses := make([]v1.NodeAddress, 0, 2)
	nodeAdresses = append(nodeAdresses, v1.NodeAddress{Type: v1.NodeInternalIP, Address: instance.PrivateIP})
	if instance.PublicIP != "" {
		nodeAdresses = append(nodeAdresses, v1.NodeAddress{Type: v1.NodeExternalIP, Address: instance.PublicIP})
	}

	return nodeAdresses
}

// NodeAddressesByProviderID returns the addresses of the specified instance.
// The instance is specified using the providerID of the node. The
// ProviderID is a unique identifier of the node. This will not be called
// from the node whose nodeaddresses are being queried. i.e. local metadata
// services cannot be used in this method to obtain nodeaddresses
func (i *instances) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]v1.NodeAddress, error) {
	instance, err := utils.CivoInstanceFromProviderID(providerID, ClusterID, i.client.civoClient)
	if err != nil {
		klog.Errorf("Unable to get instance by provider id: %s, error: %v", providerID, err)
		return nil, err
	}

	// Return Public and Private Addresses
	return getAddressessFromCivoInstance(instance), nil
}

// InstanceID returns the cloud provider ID of the node with the specified NodeName.
// Note that if the instance does not exist, we must return ("", cloudprovider.InstanceNotFound)
// cloudprovider.InstanceNotFound should NOT be returned for instances that exist but are stopped/sleeping
func (i *instances) InstanceID(ctx context.Context, nodeName types.NodeName) (string, error) {
	instance, err := utils.CivoInstanceFromName(ClusterID, string(nodeName), i.client.civoClient)
	if err != nil {
		klog.Errorf("Unable to get instance by name: %s, error: %v", string(nodeName), err)
		return "", err
	}

	return string(instance.ID), nil
}

// InstanceType returns the type of the specified instance.
func (i *instances) InstanceType(ctx context.Context, name types.NodeName) (string, error) {
	instance, err := utils.CivoInstanceFromName(ClusterID, string(name), i.client.civoClient)
	if err != nil {
		klog.Errorf("Unable to get instance by name: %s, error: %v", string(name), err)
		return "", err
	}

	return instance.Size, nil
}

// InstanceTypeByProviderID returns the type of the specified instance.
func (i *instances) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	instance, err := utils.CivoInstanceFromProviderID(providerID, ClusterID, i.client.civoClient)
	if err != nil {
		klog.Errorf("Unable to get instance by provider id: %s, error: %v", providerID, err)
		return "", err
	}

	return instance.Size, nil
}

// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all instances
// expected format for the key is standard ssh-keygen format: <protocol> <blob>
func (*instances) AddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) error {
	return nil
}

// CurrentNodeName returns the name of the node we are currently running on
// On most clouds (e.g. GCE) this is the hostname, so we provide the hostname
func (*instances) CurrentNodeName(ctx context.Context, hostname string) (types.NodeName, error) {
	return types.NodeName(hostname), nil
}

// InstanceExistsByProviderID returns true if the instance for the given provider exists.
// If false is returned with no error, the instance will be immediately deleted by the cloud controller manager.
// This method should still return true for instances that exist but are stopped/sleeping.
func (i *instances) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
	_, err := utils.CivoInstanceFromProviderID(providerID, ClusterID, i.client.civoClient)
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("Unable to get instance by provider id: %s, error: %v", providerID, err)
		return true, err
	}

	if errors.IsNotFound(err) {
		return false, nil
	}

	return true, nil
}

// InstanceShutdownByProviderID returns true if the instance is shutdown in cloudprovider
func (i *instances) InstanceShutdownByProviderID(ctx context.Context, providerID string) (bool, error) {
	instance, err := utils.CivoInstanceFromProviderID(providerID, ClusterID, i.client.civoClient)
	if err != nil {
		klog.Errorf("Unable to get instance by provider id: %s, error: %v", providerID, err)
		return false, nil
	}

	if instance.Status != statusActive {
		return true, nil
	}

	return false, nil
}
