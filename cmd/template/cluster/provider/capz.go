package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"text/template"

	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kubectl-gs/cmd/template/cluster/provider/templates/azure"
	"github.com/giantswarm/kubectl-gs/internal/key"
)

func WriteCAPZTemplate(ctx context.Context, client k8sclient.Interface, out io.Writer, config ClusterConfig) error {
	var err error

	var sshSSOPublicKey string
	{
		sshSSOPublicKey, err = key.SSHSSOPublicKey(ctx, client.CtrlClient())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var ignitionBase64 string
	{
		data := struct {
			SystemdUnits []struct {
				Name     string
				Contents string
			}
		}{
			SystemdUnits: []struct {
				Name     string
				Contents string
			}{
				{
					Name:     "set-bastion-ready.timer",
					Contents: jsonEscape(key.CapzSetBastionReadyTimer),
				},
				{
					Name:     "set-bastion-ready.service",
					Contents: jsonEscape(key.CapzSetBastionReadyService),
				},
			},
		}

		var tpl bytes.Buffer
		t := template.Must(template.New(config.FileName).Parse(fmt.Sprintf(key.BastionIgnitionTemplate, config.Name, key.BastionSSHDConfigEncoded(), base64.StdEncoding.EncodeToString([]byte(sshSSOPublicKey)))))
		err = t.Execute(&tpl, data)
		if err != nil {
			return microerror.Mask(err)
		}

		ignitionBase64 = base64.StdEncoding.EncodeToString(tpl.Bytes())
	}

	data := struct {
		BastionIgnitionSecretBase64 string
		BastionVMSize               string
		Description                 string
		KubernetesVersion           string
		Name                        string
		Namespace                   string
		SSHDConfig                  string
		SSOPublicKey                string
		Organization                string
		PodsCIDR                    string
		Version                     string
		VMSize                      string
	}{
		BastionIgnitionSecretBase64: ignitionBase64,
		BastionVMSize:               "Standard_D2_v3",
		Description:                 config.Description,
		KubernetesVersion:           "v1.19.9",
		Name:                        config.Name,
		Namespace:                   key.OrganizationNamespaceFromName(config.Organization),
		Organization:                config.Organization,
		PodsCIDR:                    config.PodsCIDR,
		SSHDConfig:                  key.NodeSSHDConfigEncoded(),
		SSOPublicKey:                sshSSOPublicKey,
		Version:                     config.ReleaseVersion,
		VMSize:                      "Standard_D4s_v3",
	}

	var templates []templateConfig
	for _, t := range azure.GetTemplates() {
		templates = append(templates, templateConfig(t))
	}

	err = runMutation(ctx, client, data, templates, out)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func jsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	// Trim the beginning and trailing " character
	return string(b[1 : len(b)-1])
}
