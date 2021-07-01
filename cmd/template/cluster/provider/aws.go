package provider

import (
	"io"
	"os"
	"text/template"

	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	capav1alpha3 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/kubectl-gs/internal/key"
	"github.com/giantswarm/kubectl-gs/internal/label"
)

func WriteAWSTemplate(out io.Writer, config ClusterCRsConfig) error {
	var err error

	if key.IsCAPAVersion(config.ReleaseVersion) {
		err = WriteCAPATemplate(out, config)
		if err != nil {
			return microerror.Mask(err)
		}
	} else {
		err = WriteGSAWSTemplate(out, config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func WriteCAPATemplate(out io.Writer, config ClusterCRsConfig) error {
	var err error

	clusterTemplate, err := getCAPAClusterTemplate(config)
	if err != nil {
		return err
	}

	data := struct {
		AWSClusterCR          string
		AWSMachineTemplateCR  string
		ClusterCR             string
		KubeadmControlPlaneCR string
	}{}

	crLabels := map[string]string{
		label.ReleaseVersion: config.ReleaseVersion,
		label.Cluster:        config.ClusterID,
		"cluster.x-k8s.io":   config.ClusterID,
		label.Organization:   config.Owner}

	objects := clusterTemplate.Objs()
	for _, o := range objects {
		switch o.GetKind() {
		case "AWSCluster":
			o.SetLabels(crLabels)
			awscluster, err := newAWSClusterFromUnstructured(config, o)
			if err != nil {
				return microerror.Mask(err)
			}
			awsClusterCRYaml, err := yaml.Marshal(awscluster)
			if err != nil {
				return microerror.Mask(err)
			}
			data.AWSClusterCR = string(awsClusterCRYaml)
		case "AWSMachineTemplate":
			o.SetLabels(crLabels)
			awsMachineTemplateCRYaml, err := yaml.Marshal(o.Object)
			if err != nil {
				return microerror.Mask(err)
			}
			data.AWSMachineTemplateCR = string(awsMachineTemplateCRYaml)
		case "Cluster":
			clusterLabels := crLabels
			for key, value := range config.Labels {
				clusterLabels[key] = value
			}
			o.SetLabels(clusterLabels)
			o.SetAnnotations(map[string]string{annotation.ClusterDescription: config.Description})
			clusterCRYaml, err := yaml.Marshal(o.Object)
			if err != nil {
				return microerror.Mask(err)
			}
			data.ClusterCR = string(clusterCRYaml)
		case "KubeadmControlPlane":
			o.SetLabels(crLabels)
			kubeadmControlPlaneCRYaml, err := yaml.Marshal(o.Object)
			if err != nil {
				return microerror.Mask(err)
			}
			data.KubeadmControlPlaneCR = string(kubeadmControlPlaneCRYaml)
		}
	}

	t := template.Must(template.New(config.FileName).Parse(key.ClusterCAPACRsTemplate))
	err = t.Execute(out, data)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func WriteGSAWSTemplate(out io.Writer, config ClusterCRsConfig) error {
	var err error

	crsConfig := v1alpha2.ClusterCRsConfig{
		ClusterID: config.ClusterID,

		ExternalSNAT:   config.ExternalSNAT,
		MasterAZ:       config.ControlPlaneAZ,
		Description:    config.Description,
		PodsCIDR:       config.PodsCIDR,
		Owner:          config.Owner,
		ReleaseVersion: config.ReleaseVersion,
		Labels:         config.Labels,
	}

	crs, err := v1alpha2.NewClusterCRs(crsConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	if config.ControlPlaneSubnet != "" {
		crs.AWSCluster.Annotations[annotation.AWSSubnetSize] = config.ControlPlaneSubnet
	}

	clusterCRYaml, err := yaml.Marshal(crs.Cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	awsClusterCRYaml, err := yaml.Marshal(crs.AWSCluster)
	if err != nil {
		return microerror.Mask(err)
	}

	g8sControlPlaneCRYaml, err := yaml.Marshal(crs.G8sControlPlane)
	if err != nil {
		return microerror.Mask(err)
	}

	awsControlPlaneCRYaml, err := yaml.Marshal(crs.AWSControlPlane)
	if err != nil {
		return microerror.Mask(err)
	}

	data := struct {
		AWSClusterCR      string
		AWSControlPlaneCR string
		ClusterCR         string
		G8sControlPlaneCR string
	}{
		AWSClusterCR:      string(awsClusterCRYaml),
		ClusterCR:         string(clusterCRYaml),
		G8sControlPlaneCR: string(g8sControlPlaneCRYaml),
		AWSControlPlaneCR: string(awsControlPlaneCRYaml),
	}

	t := template.Must(template.New(config.FileName).Parse(key.ClusterAWSCRsTemplate))
	err = t.Execute(out, data)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func getCAPAClusterTemplate(config ClusterCRsConfig) (client.Template, error) {
	var err error

	c, err := client.New("")
	if err != nil {
		return nil, err
	}

	templateOptions := client.GetClusterTemplateOptions{
		ClusterName:       config.ClusterID,
		TargetNamespace:   key.OrganizationNamespaceFromName(config.Owner),
		KubernetesVersion: "v1.19.9",
		ProviderRepositorySource: &client.ProviderRepositorySourceOptions{
			InfrastructureProvider: "aws:v0.6.6",
			Flavor:                 "machinepool",
		},
	}
	os.Setenv("AWS_SUBNET", "")
	os.Setenv("AWS_CONTROL_PLANE_MACHINE_TYPE", "")
	os.Setenv("AWS_REGION", "")
	os.Setenv("AWS_SSH_KEY_NAME", "")

	if replicas := int64(len(config.ControlPlaneAZ)); replicas > 0 {
		templateOptions.ControlPlaneMachineCount = &replicas
	}

	clusterTemplate, err := c.GetClusterTemplate(templateOptions)
	if err != nil {
		return nil, err
	}
	return clusterTemplate, nil
}

func newAWSClusterFromUnstructured(config ClusterCRsConfig, o unstructured.Unstructured) (*capav1alpha3.AWSCluster, error) {
	var awscluster capav1alpha3.AWSCluster
	{
		err := runtime.DefaultUnstructuredConverter.
			FromUnstructured(o.Object, &awscluster)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		if config.ControlPlaneAZ != nil {
			for _, az := range config.ControlPlaneAZ {
				privateSubnet := capav1alpha3.SubnetSpec{AvailabilityZone: az, IsPublic: false}
				publicSubnet := capav1alpha3.SubnetSpec{AvailabilityZone: az, IsPublic: true}
				awscluster.Spec.NetworkSpec.Subnets = append(awscluster.Spec.NetworkSpec.Subnets, &privateSubnet, &publicSubnet)
			}
		}
	}
	return &awscluster, nil
}
