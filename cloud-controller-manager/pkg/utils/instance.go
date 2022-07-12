package utils

import (
	"fmt"
	"strings"

	"github.com/civo/civogo"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
)

func civoInstanceIDFromProviderID(providerID string) (string, error) {
	if providerID == "" {
		return "", cloudprovider.InstanceNotFound
	}

	if !strings.HasPrefix(providerID, "civo://") {
		return "", fmt.Errorf("ProviderID does not match this CCM: %s", providerID)
	}

	return strings.TrimPrefix(providerID, "civo://"), nil
}

func civoInstanceFromID(clusterID, instanceID string, c civogo.Clienter) (civogo.Instance, error) {
	instance, err := c.FindKubernetesClusterInstance(clusterID, instanceID)
	if err != nil {
		klog.Errorf("Unable to find instance by id: %s, error: %v", instanceID, err)
		return civogo.Instance{}, cloudprovider.InstanceNotFound
	}

	return *instance, nil
}

// CivoInstanceFromProviderID finds civo instance by clusterID and providerID
func CivoInstanceFromProviderID(providerID, clusterID string, c civogo.Clienter) (civogo.Instance, error) {
	civoInstanceID, err := civoInstanceIDFromProviderID(providerID)
	if err != nil {
		return civogo.Instance{}, err
	}

	civoInstance, err := civoInstanceFromID(clusterID, civoInstanceID, c)
	if err != nil {
		return civogo.Instance{}, err
	}

	return civoInstance, nil
}

// CivoInstanceFromName finds civo instance by clusterID and name
func CivoInstanceFromName(clusterID, instanceName string, c civogo.Clienter) (civogo.Instance, error) {
	instance, err := c.FindKubernetesClusterInstance(clusterID, instanceName)
	if err != nil {
		klog.Errorf("Unable to find instance by name: %s, error: %v", instanceName, err)
		return civogo.Instance{}, cloudprovider.InstanceNotFound
	}

	return *instance, nil
}
