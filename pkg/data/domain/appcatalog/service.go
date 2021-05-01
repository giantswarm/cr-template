package appcatalog

import (
	"context"
	"fmt"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
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
		resource, err = s.getByName(ctx, options.Name)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	} else {
		resource, err = s.getAll(ctx)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resource, nil
}

func (s *Service) getAll(ctx context.Context) (Resource, error) {
	appCatalogCollection := &Collection{}
	var err error

	{
		selector, _ := labels.Parse("application.giantswarm.io/catalog-visibility=public")
		lo := &runtimeclient.ListOptions{
			LabelSelector: selector,
		}

		appCatalogs := &applicationv1alpha1.AppCatalogList{}
		{
			err = s.client.K8sClient.CtrlClient().List(ctx, appCatalogs, lo)
			if err != nil {
				return nil, microerror.Mask(err)
			} else if len(appCatalogs.Items) == 0 {
				return nil, microerror.Mask(noResourcesError)
			}
		}

		for _, catalog := range appCatalogs.Items {
			a := AppCatalog{
				CR: catalog.DeepCopy(),
			}
			appCatalogCollection.Items = append(appCatalogCollection.Items, a)
		}
	}

	return appCatalogCollection, nil
}

func (s *Service) getByName(ctx context.Context, name string) (Resource, error) {
	appCatalogCollection := &Collection{}
	var err error

	var selector labels.Selector

	{
		label := fmt.Sprintf("application.giantswarm.io/catalog=%s,latest=true", name)

		selector, err = labels.Parse(label)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	entries := &applicationv1alpha1.AppCatalogEntryList{}

	{
		lo := &runtimeclient.ListOptions{
			LabelSelector: selector,
		}
		err = s.client.K8sClient.CtrlClient().List(ctx, entries, lo)
		if err != nil {
			return nil, microerror.Mask(err)
		} else if len(entries.Items) == 0 {
			return nil, microerror.Mask(noResourcesError)
		}

		a := AppCatalog{
			Entries: entries,
		}
		appCatalogCollection.Items = append(appCatalogCollection.Items, a)
	}

	return appCatalogCollection, nil
}
