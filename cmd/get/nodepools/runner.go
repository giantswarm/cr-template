package nodepools

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/kubectl-gs/pkg/commonconfig"
	"github.com/giantswarm/kubectl-gs/pkg/data/domain/nodepool"
	"github.com/giantswarm/kubectl-gs/pkg/output"
)

type runner struct {
	flag   *flag
	logger micrologger.Logger
	fs     afero.Fs

	provider string
	service  nodepool.Interface

	stdout io.Writer
	stderr io.Writer
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.run(ctx, cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	var err error

	config := commonconfig.New(r.flag.config)
	{
		if r.provider == "" {
			r.provider, err = config.GetProvider()
			if err != nil {
				return microerror.Mask(err)
			}
		}

		err = r.getService(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var resource runtime.Object
	{
		options := nodepool.GetOptions{
			Provider: r.provider,
		}
		{
			if len(args) > 0 {
				options.ID = strings.ToLower(args[0])
			}

			if r.flag.AllNamespaces {
				options.Namespace = metav1.NamespaceAll
			} else {
				options.Namespace, _, err = r.flag.config.ToRawKubeConfigLoader().Namespace()
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}

		resource, err = r.service.Get(ctx, options)
		if nodepool.IsNotFound(err) {
			return microerror.Maskf(notFoundError, fmt.Sprintf("A node pool with ID '%s' cannot be found.\n", options.ID))
		} else if nodepool.IsNoResources(err) && output.IsOutputDefault(r.flag.print.OutputFormat) {
			r.printNoResourcesOutput()

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	err = r.printOutput(resource)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) getService(config *commonconfig.CommonConfig) error {
	if r.service != nil {
		return nil
	}

	client, err := config.GetClient(r.logger)
	if err != nil {
		return microerror.Mask(err)
	}

	serviceConfig := nodepool.Config{
		Client: client,
	}
	r.service, err = nodepool.New(serviceConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
