package appcatalog

import (
	"context"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// AppCatalog abstracts away the custom resource so it can be returned as a runtime
// object or a typed custom resource.
type AppCatalog struct {
	CR      *applicationv1alpha1.AppCatalog
	Entries *applicationv1alpha1.AppCatalogEntryList
}

// Collection wraps a list of appcatalogs.
type Collection struct {
	Items []AppCatalog
}

// GetOptions are the parameters that the Get method takes.
type GetOptions struct {
	LabelSelector string
	Name          string
}

type Resource interface {
	Object() runtime.Object
}

// Interface represents the contract for the appcatalog data service.
// Using this instead of a regular 'struct' makes mocking the
// service in tests much simpler.
type Interface interface {
	Get(context.Context, GetOptions) (Resource, error)
}

func (a *AppCatalog) Object() runtime.Object {
	if a.CR != nil {
		return a.CR
	}
	if a.Entries != nil {
		return a.Entries
	}

	return nil
}

func (cc *Collection) Object() runtime.Object {
	list := &metav1.List{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: "v1",
		},
		ListMeta: metav1.ListMeta{},
	}

	for _, item := range cc.Items {
		obj := item.Object()
		if obj == nil {
			continue
		}

		raw := runtime.RawExtension{
			Object: obj,
		}
		list.Items = append(list.Items, raw)
	}

	return list
}
