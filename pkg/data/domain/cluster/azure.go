package cluster

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Service) v4ListAzure(ctx context.Context) (*CommonClusterList, error) {
	var err error

	clusterConfigs := &corev1alpha1.AzureClusterConfigList{}
	{
		options := &runtimeClient.ListOptions{
			Namespace: "default",
		}
		err = s.client.K8sClient.CtrlClient().List(ctx, clusterConfigs, options)
		if err != nil {
			return nil, microerror.Mask(err)
		} else if len(clusterConfigs.Items) == 0 {
			return nil, microerror.Mask(noResourcesError)
		}
	}

	configs := &providerv1alpha1.AzureConfigList{}
	{
		options := &runtimeClient.ListOptions{
			Namespace: "default",
		}
		err = s.client.K8sClient.CtrlClient().List(ctx, configs, options)
		if err != nil {
			return nil, microerror.Mask(err)
		} else if len(clusterConfigs.Items) == 0 {
			return nil, microerror.Mask(noResourcesError)
		}
	}

	clusters := &CommonClusterList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: "v1",
		},
	}
	for _, cc := range clusterConfigs.Items {
		clusterConfig := cc

		var correspondingConfig runtime.Object
		{
			for _, config := range configs.Items {
				if cc.Name == fmt.Sprintf("%s-azure-cluster-config", config.Name) {
					correspondingConfig = &config
					break
				}
			}
			if correspondingConfig == nil {
				continue
			}
		}

		newCluster := &V4ClusterList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "List",
				APIVersion: "v1",
			},
			Items: []runtime.Object{
				&clusterConfig,
				correspondingConfig,
			},
		}

		clusters.Items = append(clusters.Items, newCluster)
	}

	return clusters, nil
}

func (s *Service) getAllAzure(ctx context.Context) ([]runtime.Object, error) {
	var (
		err      error
		clusters []runtime.Object
	)

	v4ClusterList, err := s.v4ListAzure(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	for _, c := range v4ClusterList.Items {
		clusters = append(clusters, c)
	}

	return clusters, err
}

func (s *Service) v4GetByIdAzure(ctx context.Context, id string) (*V4ClusterList, error) {
	var err error

	clusterConfig := &corev1alpha1.AzureClusterConfig{}
	{
		key := runtimeClient.ObjectKey{
			Name:      fmt.Sprintf("%s-azure-cluster-config", id),
			Namespace: "default",
		}
		err = s.client.K8sClient.CtrlClient().Get(ctx, key, clusterConfig)
		if errors.IsNotFound(err) {
			return nil, microerror.Mask(notFoundError)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	config := &providerv1alpha1.AzureConfig{}
	{
		key := runtimeClient.ObjectKey{
			Name:      id,
			Namespace: "default",
		}
		err = s.client.K8sClient.CtrlClient().Get(ctx, key, config)
		if errors.IsNotFound(err) {
			return nil, microerror.Mask(notFoundError)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	v4ClusterList := &V4ClusterList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: "v1",
		},
		Items: []runtime.Object{
			clusterConfig,
			config,
		},
	}

	return v4ClusterList, nil
}
func (s *Service) getByIdAzure(ctx context.Context, id string) (runtime.Object, error) {
	cluster, err := s.v4GetByIdAzure(ctx, id)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return cluster, nil
}
