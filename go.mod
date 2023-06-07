module github.com/civo/civo-cloud-controller-manager

go 1.16

require (
	cloud.google.com/go/compute/metadata v0.2.1 // indirect
	github.com/civo/civogo v0.3.35
	github.com/joho/godotenv v1.4.0
	github.com/onsi/gomega v1.27.7
	k8s.io/api v0.27.2
	k8s.io/apimachinery v0.27.2
	k8s.io/client-go v0.27.2
	k8s.io/cloud-provider v0.27.2
	k8s.io/component-base v0.27.2
	k8s.io/klog/v2 v2.90.1
	sigs.k8s.io/controller-runtime v0.15.0
)

replace cloud.google.com/go/compute/metadata => cloud.google.com/go/compute/metadata v0.2.1
