package cluster

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Service) v4ListKVM(ctx context.Context) (*corev1alpha1.KVMClusterConfigList, error) {
	clusters := &corev1alpha1.KVMClusterConfigList{}
	err := s.client.K8sClient.CtrlClient().List(ctx, clusters)
	if err != nil {
		return nil, microerror.Mask(err)
	} else if len(clusters.Items) == 0 {
		return nil, microerror.Mask(noResourcesError)
	}

	return clusters, nil
}

func (s *Service) v4GetByIdKVM(ctx context.Context, id string) (*corev1alpha1.KVMClusterConfig, error) {
	cluster := &corev1alpha1.KVMClusterConfig{}
	key := runtimeClient.ObjectKey{
		Name:      fmt.Sprintf("%s-kvm-cluster-config", id),
		Namespace: "default",
	}
	err := s.client.K8sClient.CtrlClient().Get(ctx, key, cluster)
	if errors.IsNotFound(err) {
		return nil, microerror.Mask(notFoundError)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	cluster.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   corev1alpha1.SchemeGroupVersion.Group,
		Version: corev1alpha1.SchemeGroupVersion.Version,
		Kind:    "KVMClusterConfig",
	})

	return cluster, nil
}

func (s *Service) getAllKVM(ctx context.Context) ([]runtime.Object, error) {
	var (
		err      error
		clusters []runtime.Object
	)

	var v4ClusterList *corev1alpha1.KVMClusterConfigList
	v4ClusterList, err = s.v4ListKVM(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	for _, c := range v4ClusterList.Items {
		clusters = append(clusters, &c)
	}

	return clusters, err
}

func (s *Service) getByIdKVM(ctx context.Context, id string) (runtime.Object, error) {
	cluster, err := s.v4GetByIdKVM(ctx, id)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return cluster, nil
}
