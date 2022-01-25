## Development

This is a quick guide on how to run CCM in development mode. 
### How to run CCM in the development

Unlike other operators, you can't as easily run CCM locally just connected in to a cluster

### A. Run CCM inside container

1. Create a tenant cluster
2. Get the kubeconfig for the tenant cluster
3. Must have a secret called `civo-api-access` within the `kube-system` namespace containing keys of `api-key`, `api-url`, `cluster-id`, `namespace` and `region`
4. Make changes, whatever you want
5. Build and push docker image
6. Create all(SA,ClusterRole,ClusterRoleBinding,Deployment) in one (Add docker image in the deployment)

    ```bash
        kubectl create -f doc/yaml/ccm-install.yaml --kubeconfig kubeconfig
    ```
### B. Run CCM from cli or outside of container

1. Create a tenant cluster
2. Get the kubeconfig for the tenant cluster
3. Set environment variables CIVO_API_URL, CIVO_API_KEY, CIVO_REGION, CIVO_CLUSTER_ID
4. Make changes, whatever you want
5. Run the CCM against the kubeconfig of the tenant cluster

   ```bash
   go run cloud-controller-manager/cmd/civo-cloud-controller-manager/main.go --kubeconfig kubeconfig
   ```

