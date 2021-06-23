package app

import (
	"context"
	"io"
	"os"
	"text/template"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/kubectl-gs/internal/key"
	templateapp "github.com/giantswarm/kubectl-gs/pkg/template/app"
)

type runner struct {
	flag   *flag
	logger micrologger.Logger
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
	var userConfigConfigMapYaml []byte
	var userConfigSecretYaml []byte
	var err error

	appConfig := templateapp.Config{
		Catalog:           r.flag.Catalog,
		Name:              r.flag.Name,
		Namespace:         r.flag.Namespace,
		Cluster:           r.flag.Cluster,
		DefaultingEnabled: r.flag.DefaultingEnabled,
		Version:           r.flag.Version,
	}

	if r.flag.flagUserSecret != "" {
		userConfigSecretData, err := key.ReadSecretYamlFromFile(afero.NewOsFs(), r.flag.flagUserSecret)
		if err != nil {
			return microerror.Mask(err)
		}

		secretConfig := templateapp.SecretConfig{
			Data:      userConfigSecretData,
			Name:      key.GenerateAssetName(r.flag.Name, "userconfig", r.flag.Cluster),
			Namespace: r.flag.Cluster,
		}
		userSecret, err := templateapp.NewSecret(secretConfig)
		if err != nil {
			return microerror.Mask(err)
		}
		appConfig.UserConfigSecretName = userSecret.GetName()

		userConfigSecretYaml, err = yaml.Marshal(userSecret)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if r.flag.flagUserConfigMap != "" {
		var configMapData string
		if r.flag.flagUserConfigMap != "" {
			configMapData, err = key.ReadConfigMapYamlFromFile(afero.NewOsFs(), r.flag.flagUserConfigMap)
			if err != nil {
				return microerror.Mask(err)
			}
		}
		configMapConfig := templateapp.ConfigMapConfig{
			Data:      configMapData,
			Name:      key.GenerateAssetName(r.flag.Name, "userconfig", r.flag.Cluster),
			Namespace: r.flag.Cluster,
		}
		userConfigMap, err := templateapp.NewConfigMap(configMapConfig)
		if err != nil {
			return microerror.Mask(err)
		}
		appConfig.UserConfigConfigMapName = userConfigMap.GetName()

		userConfigConfigMapYaml, err = yaml.Marshal(userConfigMap)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	appCRYaml, err := templateapp.NewAppCR(appConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	type AppCROutput struct {
		AppCR               string
		UserConfigSecret    string
		UserConfigConfigMap string
	}

	appCROutput := AppCROutput{
		AppCR:               string(appCRYaml),
		UserConfigConfigMap: string(userConfigConfigMapYaml),
		UserConfigSecret:    string(userConfigSecretYaml),
	}

	t := template.Must(template.New("appCR").Parse(key.AppCRTemplate))

	err = t.Execute(os.Stdout, appCROutput)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
