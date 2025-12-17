package victims

import (
	"k8s.io/client-go/dynamic"
	kube "k8s.io/client-go/kubernetes"
)

// VictimKubeClient is an adapter that exposes both typed and dynamic clients.
type VictimKubeClient interface {
	Kube() kube.Interface
	Dynamic() dynamic.Interface
}

type victimKubeClient struct {
	kubeClient    kube.Interface
	dynamicClient dynamic.Interface
}

func NewVictimClient(kubeClient kube.Interface, dynamicClient dynamic.Interface) VictimKubeClient {
	return &victimKubeClient{
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
	}
}

func (vc *victimKubeClient) Kube() kube.Interface {
	return vc.kubeClient
}

func (vc *victimKubeClient) Dynamic() dynamic.Interface {
	return vc.dynamicClient
}
