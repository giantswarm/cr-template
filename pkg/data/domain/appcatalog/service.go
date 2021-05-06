package appcatalog

import (
	"context"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kubectl-gs/pkg/data/client"
)

var _ Interface = &Service{}

// Config represent the input parameters that New takes to produce a valid appcatalog getter Service.
type Config struct {
	Client *client.Client
}

// Service is the object we'll hang the appcatalog getter methods on.
type Service struct {
	client *client.Client
}

// New returns a new appcatalog getter Service.
func New(config Config) (Interface, error) {
	if config.Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Client must not be empty", config)
	}

	s := &Service{
		client: config.Client,
	}

	return s, nil
}

// Get fetches a list of appcatalog CRs optionally filtered by name.
func (s *Service) Get(ctx context.Context, options GetOptions) (Resource, error) {
	var resource Resource
	var err error

	if len(options.Name) > 0 {
		resource, err = s.getByName(ctx, options.Name, options.LabelSelector)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		return resource, nil
	}

	resource, err = s.getAll(ctx, options.LabelSelector)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return resource, nil
}

func (s *Service) getAll(ctx context.Context, labelSelector labels.Selector) (Resource, error) {
	appCatalogCollection := &Collection{}
	var err error

	{
		lo := &runtimeclient.ListOptions{
			LabelSelector: labelSelector,
		}

		appCatalogs := &applicationv1alpha1.AppCatalogList{}
		{
			err = s.client.K8sClient.CtrlClient().List(ctx, appCatalogs, lo)
			if apimeta.IsNoMatchError(err) {
				return nil, microerror.Mask(noMatchError)
			} else if err != nil {
				return nil, microerror.Mask(err)
			} else if len(appCatalogs.Items) == 0 {
				return nil, microerror.Mask(noResourcesError)
			}
		}

		for _, appCatalog := range appCatalogs.Items {
			a := AppCatalog{
				CR: omitManagedFields(appCatalog.DeepCopy()),
			}
			appCatalogCollection.Items = append(appCatalogCollection.Items, a)
		}
	}

	return appCatalogCollection, nil
}

func (s *Service) getByName(ctx context.Context, name string, labelSelector labels.Selector) (Resource, error) {
	var err error

	appCatalogCR := &applicationv1alpha1.AppCatalog{}
	{
		err = s.client.K8sClient.CtrlClient().Get(ctx, runtimeclient.ObjectKey{
			Name: name,
		}, appCatalogCR)
		if apierrors.IsNotFound(err) {
			return nil, microerror.Mask(notFoundError)
		} else if apimeta.IsNoMatchError(err) {
			return nil, microerror.Mask(noMatchError)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	entries := &applicationv1alpha1.AppCatalogEntryList{}
	{
		lo := &runtimeclient.ListOptions{
			LabelSelector: labelSelector,
		}
		err = s.client.K8sClient.CtrlClient().List(ctx, entries, lo)
		if apimeta.IsNoMatchError(err) {
			return nil, microerror.Mask(noMatchError)
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else if len(entries.Items) == 0 {
			return nil, microerror.Mask(noResourcesError)
		}
	}

	appCatalog := &AppCatalog{
		CR:      omitManagedFields(appCatalogCR.DeepCopy()),
		Entries: entries,
	}
	appCatalog.CR.TypeMeta = metav1.TypeMeta{
		APIVersion: "appcatalog.application.giantswarm.io/v1alpha1",
		Kind:       "AppCatalog",
	}

	return appCatalog, nil
}

// omitManagedFields removes managed fields to make YAML output easier to read.
// With Kubernetes 1.21 we can use OmitManagedFieldsPrinter and remove this.
func omitManagedFields(appCatalog *applicationv1alpha1.AppCatalog) *applicationv1alpha1.AppCatalog {
	appCatalog.ManagedFields = nil
	return appCatalog
}
