package azure

import (
	_ "embed"
)

//go:embed cluster.yaml.tmpl
var cluster string

//go:embed azure_cluster.yaml.tmpl
var azureCluster string

//go:embed kubeadm_control_plane.yaml.tmpl
var kubeadmControlPlane string

//go:embed azure_machine_template.yaml.tmpl
var azureMachineTemplate string

//go:embed bastion_secret.yaml.tmpl
var bastionSecret string

//go:embed bastion_machine_deployment.yaml.tmpl
var bastionMachineDeployment string

//go:embed bastion_azure_machine_template.yaml.tmpl
var bastionAzureMachineTemplate string

type Template struct {
	Name string
	Data string
}

// GetTemplate merges all .tmpl files.
func GetTemplates() []Template {
	// Order is important here.
	// The order in this slice determines in which order files will be applied.
	return []Template{
		{Name: "cluster.yaml.tmpl", Data: cluster},
		{Name: "azure_cluster.yaml.tmpl", Data: azureCluster},
		{Name: "kubeadm_control_plane.yaml.tmpl", Data: kubeadmControlPlane},
		{Name: "azure_machine_template.yaml.tmpl", Data: azureMachineTemplate},
		{Name: "bastion_secret.yaml.tmpl", Data: bastionSecret},
		{Name: "bastion_machine_deployment.yaml.tmpl", Data: bastionMachineDeployment},
		{Name: "bastion_azure_machine_template.yaml.tmpl", Data: bastionAzureMachineTemplate},
	}
}
