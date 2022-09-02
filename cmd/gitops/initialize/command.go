package initialize

import (
	"io"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	name = "init"

	shortDescription = "Initialize GitOps repository with a basic configuration."
	longDescription  = `Initialize GitOps repository with a basic configuration.

It does not only create a basic directory structure, but also configures cloned
repository with git hooks.

The command may be run more than once, basically each user who starts working with
the GitOps repository may run it against his cloned version.

It respects the Giantswarm's GitOps repository structure recommendation:
https://github.com/giantswarm/gitops-template/blob/main/docs/repo_structure.md.`

	examples = `  # Initialize repository at the current directory
  kubectl gs gitops init

  # Initialize repository at given location
  kubectl gs gitops init --local-path /tmp/gitops-demo

  # Initialize dry-run
  kubectl gs gitops init --local-path /tmp/gitops-demo --dry-run`
)

type Config struct {
	Logger     micrologger.Logger
	FileSystem afero.Fs

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
		stderr: config.Stderr,
		stdout: config.Stdout,
	}

	c := &cobra.Command{
		Use:     name,
		Short:   shortDescription,
		Long:    longDescription,
		Example: examples,
		RunE:    r.Run,
	}

	f.Init(c)

	return c, nil
}
