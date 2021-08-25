package provider

import (
	"io"
	"text/template"

	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/kubectl-gs/internal/key"
)

func WriteAWSTemplate(out io.Writer, config ClusterCRsConfig) error {
	var err error

	if key.IsCAPAVersion(config.ReleaseVersion) {
		if config.EKS {
			err = WriteCAPAEKSTemplate(out, config)
			if err != nil {
				return microerror.Mask(err)
			}
		} else {
			err = WriteCAPATemplate(out, config)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	} else {
		err = WriteGSAWSTemplate(out, config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func WriteGSAWSTemplate(out io.Writer, config ClusterCRsConfig) error {
	var err error

	crsConfig := v1alpha3.ClusterCRsConfig{
		ClusterID: config.Name,

		ExternalSNAT:   config.ExternalSNAT,
		MasterAZ:       config.ControlPlaneAZ,
		Description:    config.Description,
		PodsCIDR:       config.PodsCIDR,
		Owner:          config.Owner,
		ReleaseVersion: config.ReleaseVersion,
		Labels:         config.Labels,
	}

	crs, err := v1alpha3.NewClusterCRs(crsConfig)
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
