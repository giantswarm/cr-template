package apps

import (
	"io"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/giantswarm/kubectl-gs/pkg/middleware"
	"github.com/giantswarm/kubectl-gs/pkg/middleware/renewtoken"
)

const (
	name  = "apps <app-name>"
	alias = "app"

	shortDescription = "Display one or many apps"
	longDescription  = `Display one or many apps

Output columns:

- NAME: Name of the app.
- APP VERSION: Version of the upstream app that is being deployed.
- VERSION: Version of the Helm chart.
- LAST DEPLOYED: When the app was last deployed.
- STATUS: Status of the app.`

	examples = `  # List all apps for the current namespace
  kubectl gs get apps
  
  # Get one app by its name
  kubectl gs get app coredns`
)

type Config struct {
	Logger     micrologger.Logger
	FileSystem afero.Fs

	K8sConfigAccess clientcmd.ConfigAccess

	Stderr io.Writer
	Stdout io.Writer
}

func New(config Config) (*cobra.Command, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.FileSystem == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.FileSystem must not be empty", config)
	}
	if config.K8sConfigAccess == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sConfigAccess must not be empty", config)
	}
	if config.Stderr == nil {
		config.Stderr = os.Stderr
	}
	if config.Stdout == nil {
		config.Stdout = os.Stdout
	}

	f := &flag{}

	r := &runner{
		flag:   f,
		logger: config.Logger,
		fs:     config.FileSystem,

		stderr: config.Stderr,
		stdout: config.Stdout,
	}

	c := &cobra.Command{
		Use:     name,
		Short:   shortDescription,
		Long:    longDescription,
		Example: examples,
		Aliases: []string{alias},
		Args:    cobra.MaximumNArgs(1),
		RunE:    r.Run,
		PreRunE: middleware.Compose(
			renewtoken.Middleware(config.K8sConfigAccess),
		),
	}

	f.Init(c)

	return c, nil
}
