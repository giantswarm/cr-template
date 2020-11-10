package provider

import (
	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	expcapiv1alpha3 "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
)

type NodePoolCRsConfig struct {
	// AWS only.
	AWSInstanceType                     string
	OnDemandBaseCapacity                int
	OnDemandPercentageAboveBaseCapacity int
	UseAlikeInstanceTypes               bool

	// Azure only.
	VMSize string

	// Common.
	FileName          string
	NodePoolID        string
	AvailabilityZones []string
	ClusterID         string
	Description       string
	NodesMax          int
	NodesMin          int
	Owner             string
	Region            string
	ReleaseComponents map[string]string
	ReleaseVersion    string
	Namespace         string
}

func newCAPIV1Alpha3MachinePoolCR(config NodePoolCRsConfig, infrastructureRef *corev1.ObjectReference) *expcapiv1alpha3.MachinePool {
	mp := &expcapiv1alpha3.MachinePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachinePool",
			APIVersion: "exp.cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.NodePoolID,
			Namespace: config.Namespace,
			Labels: map[string]string{
				label.Cluster:                 config.ClusterID,
				capiv1alpha3.ClusterLabelName: config.ClusterID,
				label.ClusterOperatorVersion:  config.ReleaseComponents["cluster-operator"],
				label.MachinePool:             config.NodePoolID,
				label.Organization:            config.Owner,
				label.ReleaseVersion:          config.ReleaseVersion,
			},
			Annotations: map[string]string{
				annotation.MachinePoolName: config.Description,
			},
		},
		Spec: expcapiv1alpha3.MachinePoolSpec{
			ClusterName:    config.ClusterID,
			Replicas:       toInt32Ptr(int32(config.NodesMax)),
			FailureDomains: config.AvailabilityZones,
			Template: capiv1alpha3.MachineTemplateSpec{
				Spec: capiv1alpha3.MachineSpec{
					ClusterName:       config.ClusterID,
					InfrastructureRef: *infrastructureRef,
				},
			},
		},
	}

	return mp
}

func newSparkCR(config NodePoolCRsConfig) *corev1alpha1.Spark {
	spark := &corev1alpha1.Spark{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Spark",
			APIVersion: "core.giantswarm.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.NodePoolID,
			Namespace: config.Namespace,
			Labels: map[string]string{
				label.Cluster:                 config.ClusterID,
				label.ReleaseVersion:          config.ReleaseVersion,
				capiv1alpha3.ClusterLabelName: config.ClusterID,
			},
		},
		Spec: corev1alpha1.SparkSpec{},
	}

	return spark
}

func toInt32Ptr(i int32) *int32 {
	return &i
}
