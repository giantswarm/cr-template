package app

import (
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

const (
	flagAppName           = "app-name"
	flagCatalog           = "catalog"
	flagCluster           = "cluster"
	flagDefaultingEnabled = "defaulting-enabled"
	flagName              = "name"
	flagNamespace         = "namespace"
	flagUserConfigMap     = "user-configmap"
	flagUserSecret        = "user-secret"
	flagVersion           = "version"
)

type flag struct {
	AppName           string
	Catalog           string
	Cluster           string
	DefaultingEnabled bool
	Name              string
	Namespace         string
	flagUserConfigMap string
	flagUserSecret    string
	Version           string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.AppName, flagAppName, "", "Optionally set a different name for the App CR.")
	cmd.Flags().StringVar(&f.Catalog, flagCatalog, "", "Catalog name where app is stored.")
	cmd.Flags().StringVar(&f.Name, flagName, "", "Name of the app in the Catalog.")
	cmd.Flags().StringVar(&f.Namespace, flagNamespace, "", "Namespace where the app will be deployed.")
	cmd.Flags().StringVar(&f.Cluster, flagCluster, "", "Name of the cluster the app will be deployed to.")
	cmd.Flags().BoolVar(&f.DefaultingEnabled, flagDefaultingEnabled, true, "Don't template fields that will be defaulted.")
	cmd.Flags().StringVar(&f.flagUserConfigMap, flagUserConfigMap, "", "Path to the user values configmap YAML file.")
	cmd.Flags().StringVar(&f.flagUserSecret, flagUserSecret, "", "Path to the user secrets YAML file.")
	cmd.Flags().StringVar(&f.Version, flagVersion, "", "App version to be installed.")
}

func (f *flag) Validate() error {
	if f.Catalog == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagCatalog)
	}
	if f.Name == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagName)
	}
	if f.Namespace == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagNamespace)
	}
	if f.Cluster == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagCluster)
	}
	if f.Version == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagVersion)
	}

	return nil
}
