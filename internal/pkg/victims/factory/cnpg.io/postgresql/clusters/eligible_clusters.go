package clusters

//All these functions require api access specific to the CRD

import (
	"context"
	"fmt"
	"strconv"

	"github.com/golang/glog"

	"kube-monkey/internal/pkg/config"
	"kube-monkey/internal/pkg/victims"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var clusterGVR = schema.GroupVersionResource{
	Group:    "postgresql.cnpg.io",
	Resource: "clusters",
	Version:  "v1",
}

func EligibleClusters(dynamicClient dynamic.Interface, namespace string, filter *metav1.ListOptions) (eligVictims []victims.Victim, err error) {
	unstructuredList, err := dynamicClient.Resource(clusterGVR).Namespace(namespace).List(context.TODO(), *filter)
	if err != nil {
		return nil, err
	}

	for _, item := range unstructuredList.Items {
		itemCopy := item
		victim, err := New(&itemCopy)
		if err != nil {
			glog.Warningf("Skipping eligible %s %s because of error: %s", item.GetKind(), item.GetName(), err.Error())
			continue
		}

		if victim.IsBlacklisted() {
			continue
		}

		eligVictims = append(eligVictims, victim)
	}

	return
}

func (c *Cluster) IsEnrolled(dynamicClient dynamic.Interface) (bool, error) {
	obj, err := dynamicClient.Resource(clusterGVR).Namespace(c.Namespace()).Get(context.TODO(), c.Name(), metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	labels := obj.GetLabels()
	return labels[config.EnabledLabelKey] == config.EnabledLabelValue, nil
}

func (c *Cluster) KillType(dynamicClient dynamic.Interface) (string, error) {
	obj, err := dynamicClient.Resource(clusterGVR).Namespace(c.Namespace()).Get(context.TODO(), c.Name(), metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	labels := obj.GetLabels()
	killType, ok := labels[config.KillTypeLabelKey]
	if !ok {
		return "", fmt.Errorf("%s %s does not have %s label", c.Kind(), c.Name(), config.KillTypeLabelKey)
	}

	return killType, nil
}

func (c *Cluster) KillValue(dynamicClient dynamic.Interface) (int, error) {
	obj, err := dynamicClient.Resource(clusterGVR).Namespace(c.Namespace()).Get(context.TODO(), c.Name(), metav1.GetOptions{})
	if err != nil {
		return -1, err
	}

	labels := obj.GetLabels()
	killMode, ok := labels[config.KillValueLabelKey]
	if !ok {
		return -1, fmt.Errorf("%s %s does not have %s label", c.Kind(), c.Name(), config.KillValueLabelKey)
	}

	killModeInt, err := strconv.Atoi(killMode)
	if err != nil || !(killModeInt > 0) {
		return -1, fmt.Errorf("Invalid value for label %s: %d", config.KillValueLabelKey, killModeInt)
	}

	return killModeInt, nil
}
