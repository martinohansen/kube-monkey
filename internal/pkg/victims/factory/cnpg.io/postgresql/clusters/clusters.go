package clusters

import (
	"fmt"
	"strconv"

	"kube-monkey/internal/pkg/config"
	"kube-monkey/internal/pkg/victims"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Cluster struct {
	*victims.VictimBase
}

// New creates a new instance of Cluster from an unstructured object
func New(obj *unstructured.Unstructured) (*Cluster, error) {
	ident, err := identifier(obj)
	if err != nil {
		return nil, err
	}
	mtbf, err := meanTimeBetweenFailures(obj)
	if err != nil {
		return nil, err
	}

	kind := obj.GetKind()
	name := obj.GetName()
	namespace := obj.GetNamespace()

	return &Cluster{VictimBase: victims.New(kind, name, namespace, ident, mtbf)}, nil
}

// Returns the value of the label defined by config.IdentLabelKey
// from the cluster labels
// This label should be unique to a cluster, and is used to
// identify the pods that belong to this cluster, as pods
// inherit labels from the Cluster
func identifier(obj *unstructured.Unstructured) (string, error) {
	labels := obj.GetLabels()
	identifier, ok := labels[config.IdentLabelKey]
	if !ok {
		return "", fmt.Errorf("%s %s does not have %s label", obj.GetKind(), obj.GetName(), config.IdentLabelKey)
	}
	return identifier, nil
}

// Read the mean-time-between-failures value defined by the Cluster
// in the label defined by config.MtbfLabelKey
func meanTimeBetweenFailures(obj *unstructured.Unstructured) (int, error) {
	labels := obj.GetLabels()
	mtbf, ok := labels[config.MtbfLabelKey]
	if !ok {
		return -1, fmt.Errorf("%s %s does not have %s label", obj.GetKind(), obj.GetName(), config.MtbfLabelKey)
	}

	mtbfInt, err := strconv.Atoi(mtbf)
	if err != nil {
		return -1, err
	}

	if !(mtbfInt > 0) {
		return -1, fmt.Errorf("Invalid value for label %s: %d", config.MtbfLabelKey, mtbfInt)
	}

	return mtbfInt, nil
}
