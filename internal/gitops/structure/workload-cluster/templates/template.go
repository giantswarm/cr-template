package workcluster

import (
	_ "embed"

	"github.com/giantswarm/kubectl-gs/v2/internal/gitops/structure/common"
)

//go:embed apps_kustomization.yaml.tmpl
var appsKustomization string

//go:embed cluster_userconfig.yaml.tmpl
var clusterUserConfig string

//go:embed default_apps_userconfig.yaml.tmpl
var defaultAppsUserConfig string

//go:embed kustomization.yaml.tmpl
var kustomization string

//go:embed patch_cluster_config.yaml.tmpl
var patchClusterConfig string

//go:embed patch_cluster_userconfig.yaml.tmpl
var patchClusterUserconfig string

//go:embed patch_default_apps_userconfig.yaml.tmpl
var patchDefaultAppsUserconfig string

//go:embed private-key.yaml.tmpl
var privateKey string

//go:embed workload-cluster.yaml.tmpl
var workloadCluster string

func GetAppsDirectoryTemplates() []common.Template {
	return []common.Template{
		common.Template{Name: "kustomization.yaml", Data: appsKustomization},
		common.Template{Name: "patch_cluster_config.yaml", Data: patchClusterConfig},
	}
}

func GetClusterDirectoryTemplates() []common.Template {
	return []common.Template{
		common.Template{Name: "kustomization.yaml", Data: kustomization},
		common.Template{Name: "cluster_userconfig.yaml", Data: clusterUserConfig},
		common.Template{Name: "default_apps_userconfig.yaml", Data: defaultAppsUserConfig},
		common.Template{Name: "patch_cluster_userconfig.yaml", Data: patchClusterUserconfig},
		common.Template{Name: "patch_default_apps_userconfig.yaml", Data: patchDefaultAppsUserconfig},
	}
}

func GetSecretsDirectoryTemplates() []common.Template {
	return []common.Template{
		common.Template{Name: "{{ .WorkloadCluster }}.gpgkey.enc.yaml", Data: privateKey},
	}
}

func GetWorkloadClusterDirectoryTemplates() []common.Template {
	return []common.Template{
		common.Template{Name: "{{ .WorkloadCluster }}.yaml", Data: workloadCluster},
	}
}
