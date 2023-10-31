package provider

import (
	"context"
	"fmt"
	"math"
	"strings"

	applicationv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kubectl-gs/v2/pkg/app"
)

type AWSConfig struct {
	ExternalSNAT       bool
	ControlPlaneSubnet string

	// for CAPA
	AWSClusterRoleIdentityName string
	MachinePool                AWSMachinePoolConfig
	NetworkAZUsageLimit        int
	NetworkVPCCIDR             string
	ClusterType                string
	HttpProxy                  string
	HttpsProxy                 string
	NoProxy                    string
	APIMode                    string
	VPCMode                    string
	TopologyMode               string
	PrefixListID               string
	TransitGatewayID           string
}

type AWSMachinePoolConfig struct {
	Name             string
	MinSize          int
	MaxSize          int
	AZs              []string
	InstanceType     string
	RootVolumeSizeGB int
	CustomNodeLabels []string
}

type AzureConfig struct {
	SubscriptionID string
}

type GCPConfig struct {
	Project           string
	FailureDomains    []string
	ControlPlane      GCPControlPlane
	MachineDeployment GCPMachineDeployment
}

type VSphereConfig struct {
	ControlPlane            VSphereControlPlane
	CredentialsSecretName   string
	ImageTemplate           string
	NetworkName             string
	Worker                  VSphereMachineTemplate
	ResourcePool            string
	ServiceLoadBalancerCIDR string
	SvcLbIpPoolName         string
}

type VSphereMachineTemplate struct {
	DiskGiB   int
	MemoryMiB int
	NumCPUs   int
	Replicas  int
}

type VSphereControlPlane struct {
	Ip         string
	IpPoolName string
	VSphereMachineTemplate
}

type GCPControlPlane struct {
	ServiceAccount ServiceAccount
}

type ServiceAccount struct {
	Email  string
	Scopes []string
}

type GCPMachineDeployment struct {
	Name             string
	FailureDomain    string
	InstanceType     string
	Replicas         int
	RootVolumeSizeGB int
	CustomNodeLabels []string
	ServiceAccount   ServiceAccount
}

type MachineConfig struct {
	BootFromVolume bool
	DiskSize       int
	Flavor         string
	Image          string
}

type OpenStackConfig struct {
	Cloud          string
	CloudConfig    string
	DNSNameservers []string

	ExternalNetworkID string
	NodeCIDR          string
	NetworkName       string
	SubnetName        string

	Bastion      MachineConfig
	ControlPlane MachineConfig
	Worker       MachineConfig

	WorkerFailureDomain string
	WorkerReplicas      int
}

type AppConfig struct {
	ClusterCatalog     string
	ClusterVersion     string
	DefaultAppsCatalog string
	DefaultAppsVersion string
}

type ClusterConfig struct {
	KubernetesVersion string
	FileName          string
	ControlPlaneAZ    []string
	Description       string
	Name              string
	Organization      string
	ReleaseVersion    string
	ReleaseComponents map[string]string
	Labels            map[string]string
	Namespace         string
	PodsCIDR          string
	OIDC              OIDC
	ServicePriority   string

	Region                   string
	BastionInstanceType      string
	BastionReplicas          int
	ControlPlaneInstanceType string

	App       AppConfig
	AWS       AWSConfig
	Azure     AzureConfig
	VSphere   VSphereConfig
	GCP       GCPConfig
	OpenStack OpenStackConfig
}

type OIDC struct {
	IssuerURL     string
	CAFile        string
	ClientID      string
	UsernameClaim string
	GroupsClaim   string
}

func newcapiClusterCR(config ClusterConfig, infrastructureRef *corev1.ObjectReference) *capi.Cluster {
	cluster := &capi.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
			Labels: map[string]string{
				label.Cluster:                config.Name,
				capi.ClusterNameLabel:        config.Name,
				label.Organization:           config.Organization,
				label.ReleaseVersion:         config.ReleaseVersion,
				label.AzureOperatorVersion:   config.ReleaseComponents["azure-operator"],
				label.ClusterOperatorVersion: config.ReleaseComponents["cluster-operator"],

				// According to RFC https://github.com/giantswarm/rfc/tree/main/classify-cluster-priority
				// we use "highest" as the default service priority.
				label.ServicePriority: config.ServicePriority,
			},
			Annotations: map[string]string{
				annotation.ClusterDescription: config.Description,
			},
		},
		Spec: capi.ClusterSpec{
			InfrastructureRef: infrastructureRef,
		},
	}

	if val, ok := config.Labels[label.ServicePriority]; ok {
		cluster.ObjectMeta.Labels[label.ServicePriority] = val
	}

	return cluster
}

func getLatestVersion(ctx context.Context, ctrlClient client.Client, app, catalog string) (string, error) {
	var catalogEntryList applicationv1alpha1.AppCatalogEntryList
	err := ctrlClient.List(ctx, &catalogEntryList, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app.kubernetes.io/name":            app,
			"application.giantswarm.io/catalog": catalog,
			"latest":                            "true",
		}),
		Namespace: "giantswarm",
	})

	if err != nil {
		return "", microerror.Mask(err)
	} else if len(catalogEntryList.Items) != 1 {
		message := fmt.Sprintf("version not specified for %s and latest release couldn't be determined in %s catalog", app, catalog)
		return "", microerror.Maskf(invalidFlagError, message)
	}

	return catalogEntryList.Items[0].Spec.Version, nil
}

func organizationNamespace(org string) string {
	return fmt.Sprintf("org-%s", org)
}

func userConfigMapName(app string) string {
	return fmt.Sprintf("%s-userconfig", app)
}

func findNextPowerOfTwo(num int) int {
	log2OfNum := math.Ceil(math.Log2(float64(num)))
	return int(math.Pow(2, log2OfNum+1))
}

func defaultTo(value string, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

// validateYAML validates the given yaml against the cluster specific app values schema
func ValidateYAML(ctx context.Context, logger micrologger.Logger, client k8sclient.Interface, clusterApp applicationv1alpha1.App, yaml map[string]interface{}) error {

	serviceConfig := app.Config{
		Client: client,
		Logger: logger,
	}
	service, err := app.New(serviceConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	_, resultJsonValidate, err := service.ValidateApp(ctx, &clusterApp, "", yaml)
	if err != nil {
		return microerror.Mask(err)
	}

	resultErrors := resultJsonValidate.Errors()
	var validationErrors []string
	if len(resultErrors) > 0 {
		for _, resultError := range resultErrors {
			validationErrors = append(validationErrors, fmt.Errorf("%s", resultError.Description()).Error())
		}
		// return all validation errors
		return microerror.Mask(fmt.Errorf(strings.Join(validationErrors, "; ")))
	}

	return nil
}
