/*
Package kubernetes is the km k8 package that sets up the configured k8 clientset used to communicate with the apiserver

Use CreateClient to create and verify connectivity.
It's recommended to create a new clientset after a period of inactivity
*/
package kubernetes

import (
	"fmt"
	"path/filepath"

	"github.com/golang/glog"

	cfg "kube-monkey/internal/pkg/config"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// CreateClient creates, verifies and returns an instance of k8 clientset
func CreateClient() (*kube.Clientset, dynamic.Interface, error) {
	client, dynamicClient, err := NewClusterClient()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate NewInClusterClient: %v", err)
	}

	if VerifyClient(client) {
		return client, dynamicClient, nil
	}
	return nil, nil, fmt.Errorf("Unable to verify client connectivity to Kubernetes apiserver")
}

// NewClusterClient only creates an initialized instance of k8 clientset
func NewClusterClient() (*kube.Clientset, dynamic.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		if err == rest.ErrNotInCluster {
			// Attempt to use out of cluster config
			config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
			if err != nil {
				glog.Errorf("failed to obtain config from kubeconfig file: %v", err)
				return nil, nil, err
			}
		} else {
			glog.Errorf("failed to obtain config from InClusterConfig: %v", err)
			return nil, nil, err
		}
	}

	if apiserverHost, override := cfg.ClusterAPIServerHost(); override {
		glog.V(5).Infof("API server host overridden to: %s\n", apiserverHost)
		config.Host = apiserverHost
	}

	clientset, err := kube.NewForConfig(config)
	if err != nil {
		glog.Errorf("failed to create clientset in NewForConfig: %v", err)
		return nil, nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		glog.Errorf("failed to create dynamic client: %v", err)
		return nil, nil, err
	}
	return clientset, dynamicClient, nil
}

func VerifyClient(client discovery.DiscoveryInterface) bool {
	_, err := client.ServerVersion()
	return err == nil
}
