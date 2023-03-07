module github.com/civo/civo-cloud-controller-manager

go 1.16

require (
	github.com/civo/civogo v0.3.19
	github.com/emicklei/go-restful v2.16.0+incompatible // indirect
	github.com/go-openapi/swag v0.19.7 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/joho/godotenv v1.4.0
	github.com/onsi/gomega v1.19.0
	github.com/sirupsen/logrus v1.8.1 // indirect
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cloud-provider v0.21.1
	k8s.io/component-base v0.21.1
	k8s.io/klog/v2 v2.9.0
	sigs.k8s.io/controller-runtime v0.0.0-00010101000000-000000000000
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
	github.com/openshift/api => github.com/openshift/api v0.0.0-20200526144822-34f54f12813a
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20200521150516-05eb9880269c
	github.com/operator-framework/operator-lifecycle-manager => github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190128024246-5eb7ae5bdb7a
	k8s.io/client-go => k8s.io/client-go v0.21.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.1
	k8s.io/kubectl => k8s.io/kubectl v0.21.1
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.9.0
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.0
)
