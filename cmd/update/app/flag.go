package app

import (
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	flagVersion = "version"
	flagAppName = "name"
)

type flag struct {
	config  genericclioptions.RESTClientGetter
	print   *genericclioptions.PrintFlags
	Version string
	Name    string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.Version, flagVersion, "", "Version to update the app to")
	// Hide flag in favour of the longDescription, otherwise if the number of supported
	// update flags grows, it may be hard to differentiate them from the rest of the flags,
	// like kubectl global flags.
	_ = cmd.Flags().MarkHidden(flagVersion)

	cmd.Flags().StringVar(&f.Name, flagAppName, "", "Name of the app to update")
	_ = cmd.MarkFlagRequired(flagAppName)

	f.config = genericclioptions.NewConfigFlags(true)
	f.print = genericclioptions.NewPrintFlags("")

	// Merging current command flags and config flags,
	// to be able to override kubectl-specific ones.
	f.config.(*genericclioptions.ConfigFlags).AddFlags(cmd.Flags())
	f.print.AddFlags(cmd)
}

func (f *flag) Validate() error {
	// at least one of the supported updates must be provided
	if len(f.Version) > 0 {
		return nil
	}

	return microerror.Maskf(notEnoughFlags, "at least one of the --version parameters has to be provided")
}
