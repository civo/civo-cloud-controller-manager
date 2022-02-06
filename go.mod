module github.com/civo/civo-cloud-controller-manager

go 1.16

require (
	github.com/civo/civogo v0.2.66
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/go-openapi/swag v0.19.7 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/onsi/gomega v1.13.0
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/net v0.0.0-20210928044308-7d9f5e0b762b // indirect
	google.golang.org/appengine v1.6.7 // indirect
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cloud-provider v0.21.1
	k8s.io/component-base v0.21.1
	k8s.io/klog/v2 v2.9.0
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b // indirect
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
