package civo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/civo/civo-cloud-controller-manager/cloud-controller-manager/pkg/utils"
	"github.com/civo/civogo"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	cloudprovider "k8s.io/cloud-provider"
	cloudproviderapi "k8s.io/cloud-provider/api"
	"k8s.io/klog/v2"
)

type instancesV2 struct {
	client *clients
}

func newInstancesV2(c *clients) cloudprovider.InstancesV2 {
	return &instancesV2{
		client: c,
	}
}

// findInstance looks the node's instance up by providerID, falling back to the
// node name when the providerID has not been set yet (first initialization).
func (i *instancesV2) findInstance(node *v1.Node) (civogo.Instance, error) {
	if node.Spec.ProviderID != "" {
		return utils.CivoInstanceFromProviderID(node.Spec.ProviderID, ClusterID, i.client.civoClient)
	}
	return utils.CivoInstanceFromName(ClusterID, node.Name, i.client.civoClient)
}

// InstanceExists returns true if the instance for the given node exists.
// If false is returned with no error, the node will be deleted by the cloud controller manager.
func (i *instancesV2) InstanceExists(ctx context.Context, node *v1.Node) (bool, error) {
	_, err := i.findInstance(node)
	if errors.Is(err, cloudprovider.InstanceNotFound) {
		return false, nil
	}
	if err != nil {
		klog.Errorf("Unable to get instance for node %s, error: %v", node.Name, err)
		return true, err
	}
	return true, nil
}

// InstanceShutdown returns true if the instance for the given node is shut down.
func (i *instancesV2) InstanceShutdown(ctx context.Context, node *v1.Node) (bool, error) {
	instance, err := i.findInstance(node)
	if err != nil {
		klog.Errorf("Unable to get instance for node %s, error: %v", node.Name, err)
		return false, nil
	}
	return instance.Status != statusActive, nil
}

// InstanceMetadata returns the instance's metadata.
//
// While the node still carries the cloud provider's uninitialized taint, the
// labels and taints of the node's pool are applied to the node first. The cloud
// node controller removes the uninitialized taint only after this returns
// successfully, so the node never becomes schedulable without its pool's
// labels and taints. Any failure is returned so the controller retries and the
// node stays unschedulable.
func (i *instancesV2) InstanceMetadata(ctx context.Context, node *v1.Node) (*cloudprovider.InstanceMetadata, error) {
	instance, err := i.findInstance(node)
	if err != nil {
		klog.Errorf("Unable to get instance for node %s, error: %v", node.Name, err)
		return nil, err
	}

	if hasUninitializedTaint(node) {
		if err := i.applyPoolLabelsAndTaints(ctx, node.Name, instance.ID); err != nil {
			return nil, fmt.Errorf("applying pool labels and taints to node %s: %w", node.Name, err)
		}
	}

	return &cloudprovider.InstanceMetadata{
		ProviderID:    fmt.Sprintf("%s://%s", ProviderName, instance.ID),
		InstanceType:  instance.Size,
		NodeAddresses: getAddressessFromCivoInstance(instance),
	}, nil
}

func (i *instancesV2) applyPoolLabelsAndTaints(ctx context.Context, nodeName, instanceID string) error {
	if i.client.kclient == nil {
		return errors.New("kubernetes client is not initialized")
	}

	cluster, err := i.client.civoClient.GetKubernetesCluster(ClusterID)
	if err != nil {
		return fmt.Errorf("getting cluster %s: %w", ClusterID, err)
	}

	pool := findPoolForNode(cluster, nodeName, instanceID)
	if pool == nil {
		return fmt.Errorf("no pool in cluster %s contains node %s", ClusterID, nodeName)
	}

	nodes := i.client.kclient.CoreV1().Nodes()
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		node, err := nodes.Get(ctx, nodeName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		labelsChanged := mergeLabels(node, pool.Labels)
		taintsChanged := mergeTaints(node, pool.Taints)
		if !labelsChanged && !taintsChanged {
			return nil
		}

		if _, err := nodes.Update(ctx, node, metav1.UpdateOptions{}); err != nil {
			return err
		}
		klog.Infof("Applied labels and taints of pool %s to node %s", pool.ID, nodeName)
		return nil
	})
}

func hasUninitializedTaint(node *v1.Node) bool {
	for _, t := range node.Spec.Taints {
		if t.Key == cloudproviderapi.TaintExternalCloudProvider {
			return true
		}
	}
	return false
}

func findPoolForNode(cluster *civogo.KubernetesCluster, nodeName, instanceID string) *civogo.KubernetesPool {
	for idx := range cluster.Pools {
		pool := &cluster.Pools[idx]
		for _, name := range pool.InstanceNames {
			if strings.EqualFold(name, nodeName) {
				return pool
			}
		}
		for _, inst := range pool.Instances {
			if inst.ID == instanceID || strings.EqualFold(inst.Hostname, nodeName) {
				return pool
			}
		}
	}
	return nil
}

func mergeLabels(node *v1.Node, labels map[string]string) bool {
	changed := false
	for k, v := range labels {
		if node.Labels == nil {
			node.Labels = map[string]string{}
		}
		if cur, ok := node.Labels[k]; !ok || cur != v {
			node.Labels[k] = v
			changed = true
		}
	}
	return changed
}

func mergeTaints(node *v1.Node, taints []v1.Taint) bool {
	changed := false
	for _, want := range taints {
		found := false
		for idx := range node.Spec.Taints {
			cur := &node.Spec.Taints[idx]
			if cur.Key == want.Key && cur.Effect == want.Effect {
				found = true
				if cur.Value != want.Value {
					cur.Value = want.Value
					changed = true
				}
				break
			}
		}
		if !found {
			node.Spec.Taints = append(node.Spec.Taints, v1.Taint{Key: want.Key, Value: want.Value, Effect: want.Effect})
			changed = true
		}
	}
	return changed
}
