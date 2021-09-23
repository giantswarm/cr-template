package provider

import (
	"context"
	"io"

	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kubectl-gs/cmd/template/cluster/provider/templates/vsphere"
	"github.com/giantswarm/kubectl-gs/internal/key"
)

func WriteCAPVTemplate(ctx context.Context, client k8sclient.Interface, out io.Writer, config ClusterCRsConfig) error {
	var err error

	data := struct {
		Description       string
		KubernetesVersion string
		Name              string
		Namespace         string
		Organization      string
		Version           string
	}{
		Description:       config.Description,
		KubernetesVersion: "v1.19.9",
		Name:              config.Name,
		Namespace:         key.OrganizationNamespaceFromName(config.Organization),
		Organization:      config.Organization,
		Version:           config.ReleaseVersion,
	}

	err = runMutation(ctx, client, data, vsphere.GetTemplates(), out)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
