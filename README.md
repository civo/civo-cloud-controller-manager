# Civo Kubernetes Cloud Controller Manager

This controller is the Kubernetes cloud controller manager implementation for Civo. This controller is installed in to Civo K3s client clusters and handles the all operations related to Civo LoadBalancer.

## Getting Started

How do I run the `civo-cloud-controller-manager` [here!](https://github.com/civo/civo-cloud-controller-manager/blob/master/doc/getting-started.md)

## Load Balancers

The CCM will listen for Services with a type `LoadBalancer` set and provision an external load balancer within the Civo platform. 

Once the Load Balancer is provisioned, a public DNS entry will be added mapping to the id of the load balancer. e.g. 92f8162c-c23c-4019-b6a0-2c18b8363f50.lb.civo.com.

Read More: https://www.civo.com/learn/managing-external-load-balancers-on-civo

### Load Balancer Customisation

| Annotation | Description | Example Values |
|------------|-------------|----------------|
| kubernetes.civo.com/firewall-id | If provided, an existing Firewall will be used. | 03093EF6-31E6-48B1-AB1D-152AC3A8C90A |
| kubernetes.civo.com/max-concurrent-requests | If provided, the user can specify the max number of concurrent requests a loadbalancer could afford. Note: Defaults to 10000. Whenever user updates this value, there will be a downtime for the loadbalancer (usually < 1 minute)  | 10000, 20000, 30000, 40000, 50000 |
| kubernetes.civo.com/loadbalancer-enable-proxy-protocol | If set, a proxy-protocol header will be sent via the load balancer. <br /><br />NB: This requires support from the Service End Points within the cluster. | send-proxy<br />send-proxy-v2 |
| kubernetes.civo.com/loadbalancer-algorithm | Custom the algorithm the external load balancer uses | round_robin<br />least_connections |
| kubernetes.civo.com/ipv4-address | If set, LoadBalancer will have the mentioned IP as the public IP. Please note: the reserved IP should be present in the account before claiming it. | 10.0.0.20<br/> my-reserved-ip |
| kubernetes.civo.com/protocol | If set, this will override the protocol set on the svc with this | http<br />tcp |
| kubernetes.civo.com/server-timeout| server timeout determines how long the load balancer will wait for a response from the server/upstream. Defaults to 60s | 60s<br /> 120m |
| kubernetes.civo.com/client-timeout| client timeout determines how long the load balancer will wait for any activity from the client/downstream. Defaults to 60s | 60s<br /> 120m |

### Load Balancer Status Annotations

| Annotation                            | Description                                            | Sample Value                         |
| ------------------------------------- | ------------------------------------------------------ | ------------------------------------ |
| kubernetes.civo.com/cluster-id        | The ID of the cluster the load balancer is assigned to | 05CE1CA2-067F-42F0-9BAA-17A6A800EFBB |
| kubernetes.civo.com/loadbalancer-id   | The ID of the Load Balancer within Civo.               | 92F8162C-C23C-4019-B6A0-2C18B8363F50 |
| kubernetes.civo.com/loadbalancer-name | The name of the Load Balancer                          | Lb-test                              |



### Proxy Protocol

When the proxy protocol annotation (`kubernetes.civo.com/loadbalancer-enable-proxy-protocol`) is set, the IP of the LoadBalancer is not set, only the Hostname. This will mean that all traffic local to the cluster to the LoadBalancer end point is now sent via the LoadBalancer. This allows services like CertManager to work correctly.

This option is currently a workaround for the issue https://github.com/kubernetes/ingress-nginx/issues/3996, should be removed or refactored after the Kubernetes [KEP-1860]


## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/civo/civo-cloud-controller-manager