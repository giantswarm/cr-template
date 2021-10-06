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

//go:embed bastion.yaml.tmpl
var bastion string

// GetTemplate merges all .tmpl files.
func GetTemplates() []string {
	// Order is important here.
	// The order in this slice determines in which order files will be applied.
	return []string{
		cluster,
		azureCluster,
		kubeadmControlPlane,
		azureMachineTemplate,
		bastion,
	}
}
