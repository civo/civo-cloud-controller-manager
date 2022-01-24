package civo

import (
	"github.com/civo/civogo"
	"k8s.io/client-go/kubernetes"
)

type clients struct {
	civoClient *civogo.Client
	kclient    kubernetes.Interface
}

func newClients(client *civogo.Client) *clients {
	return &clients{
		civoClient: client,
	}
}
