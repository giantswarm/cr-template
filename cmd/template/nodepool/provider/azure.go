package provider

import (
	"fmt"
	"io"
	"strconv"
	"text/template"

	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/reference"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/kubectl-gs/internal/key"
)

func WriteAzureTemplate(out io.Writer, config NodePoolCRsConfig) error {
	var err error

	var azureMachinePoolCRYaml, machinePoolCRYaml, sparkCRYaml []byte
	{
		azureMachinePoolCR := newAzureMachinePoolCR(config)
		azureMachinePoolCRYaml, err = yaml.Marshal(azureMachinePoolCR)
		if err != nil {
			return microerror.Mask(err)
		}

		sparkCR := newSparkCR(config)
		sparkCRYaml, err = yaml.Marshal(sparkCR)
		if err != nil {
			return microerror.Mask(err)
		}

		infrastructureRef := newCAPZMachinePoolInfraRef(azureMachinePoolCR)
		bootstrapConfigRef := newSparkCRRef(sparkCR)

		machinePoolCR := newCAPIV1Alpha3MachinePoolCR(config, infrastructureRef, bootstrapConfigRef)
		{
			machinePoolCR.GetAnnotations()[annotation.NodePoolMinSize] = strconv.Itoa(config.NodesMin)
			machinePoolCR.GetAnnotations()[annotation.NodePoolMaxSize] = strconv.Itoa(config.NodesMax)
		}
		machinePoolCRYaml, err = yaml.Marshal(machinePoolCR)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	data := struct {
		ProviderMachinePoolCR string
		MachinePoolCR         string
		SparkCR               string
	}{
		ProviderMachinePoolCR: string(azureMachinePoolCRYaml),
		MachinePoolCR:         string(machinePoolCRYaml),
		SparkCR:               string(sparkCRYaml),
	}

	t := template.Must(template.New(config.FileName).Parse(key.MachinePoolAzureCRsTemplate))
	err = t.Execute(out, data)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func newAzureMachinePoolCR(config NodePoolCRsConfig) *expcapzv1alpha3.AzureMachinePool {
	var spot *v1alpha3.SpotVMOptions
	if config.AzureUseSpotVms {
		var maxPrice resource.Quantity
		if config.AzureSpotMaxPrice > 0 {
			maxPrice = resource.MustParse(fmt.Sprintf("%f", config.AzureSpotMaxPrice))

		} else {
			maxPrice = resource.MustParse("-1")
		}
		spot = &v1alpha3.SpotVMOptions{
			MaxPrice: &maxPrice,
		}
	}

	azureMp := &expcapzv1alpha3.AzureMachinePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureMachinePool",
			APIVersion: "exp.infrastructure.cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.NodePoolID,
			Namespace: config.Namespace,
			Labels: map[string]string{
				label.Cluster:                 config.ClusterName,
				capiv1alpha3.ClusterLabelName: config.ClusterName,
				label.MachinePool:             config.NodePoolID,
				label.Organization:            config.Owner,
			},
		},
		Spec: expcapzv1alpha3.AzureMachinePoolSpec{
			Template: expcapzv1alpha3.AzureMachineTemplate{
				VMSize:        config.VMSize,
				SSHPublicKey:  "",
				SpotVMOptions: spot,
			},
		},
	}

	return azureMp
}

func newCAPZMachinePoolInfraRef(obj runtime.Object) *corev1.ObjectReference {
	var infrastructureCRRef *corev1.ObjectReference
	{
		s := runtime.NewScheme()
		err := expcapzv1alpha3.AddToScheme(s)
		if err != nil {
			panic(fmt.Sprintf("expcapzv1alpha3.AddToScheme: %+v", err))
		}

		infrastructureCRRef, err = reference.GetReference(s, obj)
		if err != nil {
			panic(fmt.Sprintf("cannot create reference to infrastructure CR: %q", err))
		}
	}

	return infrastructureCRRef
}

func newSparkCRRef(obj runtime.Object) *corev1.ObjectReference {
	var infrastructureCRRef *corev1.ObjectReference
	{
		s := runtime.NewScheme()
		err := corev1alpha1.AddToScheme(s)
		if err != nil {
			panic(fmt.Sprintf("corev1alpha1.AddToScheme: %+v", err))
		}

		infrastructureCRRef, err = reference.GetReference(s, obj)
		if err != nil {
			panic(fmt.Sprintf("cannot create reference to infrastructure CR: %q", err))
		}
	}

	return infrastructureCRRef
}
