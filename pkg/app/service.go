package app

import (
	"github.com/giantswarm/app/v6/pkg/values"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appdata "github.com/giantswarm/kubectl-gs/v2/pkg/data/domain/app"
	catalogdata "github.com/giantswarm/kubectl-gs/v2/pkg/data/domain/catalog"
	"github.com/giantswarm/kubectl-gs/v2/pkg/helmbinary"
)

var _ Interface = &Service{}

// Config represents the values that New() needs in order to create a valid app service.
type Config struct {
	Client k8sclient.Interface
	Logger micrologger.Logger
}

// Service represents an instance of the App service.
type Service struct {
	client             client.Client
	appDataService     appdata.Interface
	catalogDataService catalogdata.Interface
	helmbinaryService  helmbinary.Interface
	valuesService      *values.Values

	catalogFetchResults map[string]CatalogFetchResult
	schemaFetchResults  map[string]SchemaFetchResult
}

// New returns an app service given a certain Config.
func New(config Config) (Interface, error) {
	var err error

	if config.Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Client must not be empty", config)
	}

	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	var appDataService appdata.Interface
	{
		c := appdata.Config{
			Client: config.Client.CtrlClient(),
		}

		appDataService, err = appdata.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var catalogDataService catalogdata.Interface
	{
		c := catalogdata.Config{
			Client: config.Client.CtrlClient(),
		}

		catalogDataService, err = catalogdata.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var helmbinaryService helmbinary.Interface
	{
		c := helmbinary.Config{
			Client: config.Client.CtrlClient(),
		}

		helmbinaryService, err = helmbinary.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var valuesService *values.Values
	{
		c := values.Config{
			K8sClient: config.Client.K8sClient(),
			Logger:    config.Logger,
		}

		valuesService, err = values.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		client:              config.Client.CtrlClient(),
		appDataService:      appDataService,
		catalogDataService:  catalogDataService,
		helmbinaryService:   helmbinaryService,
		valuesService:       valuesService,
		catalogFetchResults: make(map[string]CatalogFetchResult),
		schemaFetchResults:  make(map[string]SchemaFetchResult),
	}

	return s, nil
}
