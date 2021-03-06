package civo

import (
	"github.com/civo/civogo"
	"k8s.io/client-go/kubernetes"
)

type clients struct {
	civoClient civogo.Clienter
	kclient    kubernetes.Interface
}

func newClients(client civogo.Clienter) *clients {
	return &clients{
		civoClient: client,
	}
}
