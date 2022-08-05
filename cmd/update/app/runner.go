package app

import (
	"context"
	"fmt"
	"io"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/giantswarm/kubectl-gs/pkg/commonconfig"
	"github.com/giantswarm/kubectl-gs/pkg/data/domain/app"
)

type runner struct {
	commonConfig *commonconfig.CommonConfig
	configFlags  *genericclioptions.RESTClientGetter
	flag         *flag
	logger       micrologger.Logger

	service app.Interface

	stdout io.Writer
	stderr io.Writer
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.flag.Validate()
	if err != nil {
		_ = cmd.Help()
		return microerror.Mask(err)
	}

	r.commonConfig = commonconfig.New(*r.configFlags)
	err = r.run(ctx, cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	var err error

	{
		err = r.getService(r.commonConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	namespace, _, err := r.commonConfig.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return microerror.Mask(err)
	}

	name := r.flag.Name
	version := r.flag.Version

	patchOptions := app.PatchOptions{
		Namespace: namespace,
		Name:      name,
		Version:   version,
	}

	err = r.service.Patch(ctx, patchOptions)
	if app.IsNotFound(err) {
		return microerror.Maskf(notFoundError, "An app with name '%s' cannot be found in the '%s' namespace.\n", patchOptions.Name, patchOptions.Namespace)
	} else if app.IsNoResources(err) {
		return microerror.Maskf(noResourcesError, "No app with the name '%s' and the version '%s' found in the catalog.\n", patchOptions.Name, patchOptions.Version)
	} else if err != nil {
		return microerror.Mask(err)
	}

	fmt.Fprintf(r.stdout, "App '%s' updated to version '%s'\n", name, version)
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

	serviceConfig := app.Config{
		Client: client.CtrlClient(),
	}
	r.service, err = app.New(serviceConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
