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

func identifier(obj *unstructured.Unstructured) (string, error) {
	labels := obj.GetLabels()
	identifier, ok := labels[config.IdentLabelKey]
	if !ok {
		return "", fmt.Errorf("%s %s does not have %s label", obj.GetKind(), obj.GetName(), config.IdentLabelKey)
	}
	return identifier, nil
}

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
		return -1, fmt.Errorf("invalid value for label %s: %d", config.MtbfLabelKey, mtbfInt)
	}

	return mtbfInt, nil
}
