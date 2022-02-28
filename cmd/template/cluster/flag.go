package cluster

import (
	"net"
	"regexp"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/giantswarm/kubectl-gs/cmd/template/cluster/provider"
	"github.com/giantswarm/kubectl-gs/internal/key"
	"github.com/giantswarm/kubectl-gs/pkg/labels"
)

const (
	flagProvider = "provider"

	// AWS only.
	flagAWSExternalSNAT       = "external-snat"
	flagAWSEKS                = "aws-eks"
	flagAWSControlPlaneSubnet = "control-plane-subnet"

	// App-based clusters only.
	flagClusterCatalog     = "cluster-catalog"
	flagClusterVersion     = "cluster-version"
	flagDefaultAppsCatalog = "default-apps-catalog"
	flagDefaultAppsVersion = "default-apps-version"

	// OpenStack only.
	flagOpenStackCloud                      = "cloud"
	flagOpenStackCloudConfig                = "cloud-config"
	flagOpenStackDNSNameservers             = "dns-nameservers"
	flagOpenStackExternalNetworkID          = "external-network-id"
	flagOpenStackNodeCIDR                   = "node-cidr"
	flagOpenStackBastionBootFromVolume      = "bastion-boot-from-volume"
	flagOpenStackBastionDiskSize            = "bastion-disk-size"
	flagOpenStackBastionImage               = "bastion-image"
	flagOpenStackBastionMachineFlavor       = "bastion-machine-flavor"
	flagOpenStackControlPlaneBootFromVolume = "control-plane-boot-from-volume"
	flagOpenStackControlPlaneDiskSize       = "control-plane-disk-size"
	flagOpenStackControlPlaneImage          = "control-plane-image"
	flagOpenStackControlPlaneMachineFlavor  = "control-plane-machine-flavor"
	flagOpenStackWorkerBootFromVolume       = "worker-boot-from-volume"
	flagOpenStackWorkerDiskSize             = "worker-disk-size"
	flagOpenStackWorkerFailureDomain        = "worker-failure-domain"
	flagOpenStackWorkerImage                = "worker-image"
	flagOpenStackWorkerMachineFlavor        = "worker-machine-flavor"
	flagOpenStackWorkerReplicas             = "worker-replicas"

	// Common.
	flagControlPlaneAZ    = "control-plane-az"
	flagDescription       = "description"
	flagKubernetesVersion = "kubernetes-version"
	flagName              = "name"
	flagOIDCIssuerURL     = "oidc-issuer-url"
	flagOIDCCAFile        = "oidc-ca-file"
	flagOIDCClientID      = "oidc-client-id"
	flagOIDCUsernameClaim = "oidc-username-claim"
	flagOIDCGroupsClaim   = "oidc-groups-claim"
	flagOutput            = "output"
	flagOrganization      = "organization"
	flagPodsCIDR          = "pods-cidr"
	flagRelease           = "release"
	flagLabel             = "label"
)

type flag struct {
	Provider string

	// Common.
	ControlPlaneAZ    []string
	Description       string
	KubernetesVersion string
	Name              string
	Output            string
	Organization      string
	PodsCIDR          string
	Release           string
	Label             []string

	// Provider-specific
	AWS       provider.AWSConfig
	OpenStack provider.OpenStackConfig
	App       provider.AppConfig
	OIDC      provider.OIDC

	config genericclioptions.RESTClientGetter
	print  *genericclioptions.PrintFlags
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.Provider, flagProvider, "", "Installation infrastructure provider.")

	// AWS only.
	cmd.Flags().StringVar(&f.AWS.ControlPlaneSubnet, flagAWSControlPlaneSubnet, "", "Subnet used for the Control Plane.")
	cmd.Flags().BoolVar(&f.AWS.ExternalSNAT, flagAWSExternalSNAT, false, "AWS CNI configuration.")
	cmd.Flags().BoolVar(&f.AWS.EKS, flagAWSEKS, false, "Enable AWSEKS. Only available for AWS Release v20.0.0 (CAPA)")

	// OpenStack only.
	cmd.Flags().StringVar(&f.OpenStack.Cloud, flagOpenStackCloud, "", "Name of cloud (OpenStack only).")
	cmd.Flags().StringVar(&f.OpenStack.CloudConfig, flagOpenStackCloudConfig, "", "Name of cloud config (OpenStack only).")
	cmd.Flags().StringSliceVar(&f.OpenStack.DNSNameservers, flagOpenStackDNSNameservers, nil, "DNS nameservers (OpenStack only).")
	cmd.Flags().StringVar(&f.OpenStack.ExternalNetworkID, flagOpenStackExternalNetworkID, "", "External network ID (OpenStack only).")
	cmd.Flags().StringVar(&f.OpenStack.NodeCIDR, flagOpenStackNodeCIDR, "", "CIDR used for the nodes (OpenStack only).")
	// bastion
	cmd.Flags().BoolVar(&f.OpenStack.Bastion.BootFromVolume, flagOpenStackBastionBootFromVolume, false, "Bastion boot from volume (OpenStack only).")
	cmd.Flags().IntVar(&f.OpenStack.Bastion.DiskSize, flagOpenStackBastionDiskSize, 10, "Bastion machine root volume disk size (OpenStack only).")
	cmd.Flags().StringVar(&f.OpenStack.Bastion.Image, flagOpenStackBastionImage, "", "Bastion machine image (OpenStack only).")
	cmd.Flags().StringVar(&f.OpenStack.Bastion.Flavor, flagOpenStackBastionMachineFlavor, "", "Bastion machine flavor (OpenStack only).")
	// control plane
	cmd.Flags().BoolVar(&f.OpenStack.ControlPlane.BootFromVolume, flagOpenStackControlPlaneBootFromVolume, false, "Control plane boot from volume (OpenStack only).")
	cmd.Flags().IntVar(&f.OpenStack.ControlPlane.DiskSize, flagOpenStackControlPlaneDiskSize, 0, "Control plane machine root volume disk size (OpenStack only).")
	cmd.Flags().StringVar(&f.OpenStack.ControlPlane.Image, flagOpenStackControlPlaneImage, "", "Control plane machine image (OpenStack only).")
	cmd.Flags().StringVar(&f.OpenStack.ControlPlane.Flavor, flagOpenStackControlPlaneMachineFlavor, "", "Control plane machine flavor (OpenStack only).")
	// workers
	cmd.Flags().BoolVar(&f.OpenStack.Worker.BootFromVolume, flagOpenStackWorkerBootFromVolume, false, "Default worker node pool boot from volume (OpenStack only).")
	cmd.Flags().IntVar(&f.OpenStack.Worker.DiskSize, flagOpenStackWorkerDiskSize, 0, "Default worker node pool machine root volume disk size (OpenStack only).")
	cmd.Flags().StringVar(&f.OpenStack.WorkerFailureDomain, flagOpenStackWorkerFailureDomain, "", "Default worker node pool failure domain (OpenStack only).")
	cmd.Flags().StringVar(&f.OpenStack.Worker.Image, flagOpenStackWorkerImage, "", "Default worker node pool machine image name (OpenStack only).")
	cmd.Flags().StringVar(&f.OpenStack.Worker.Flavor, flagOpenStackWorkerMachineFlavor, "", "Default worker node pool machine flavor (OpenStack only).")
	cmd.Flags().IntVar(&f.OpenStack.WorkerReplicas, flagOpenStackWorkerReplicas, 0, "Default worker node pool replicas (OpenStack only).")

	// App-based clusters only.
	cmd.Flags().StringVar(&f.App.ClusterCatalog, flagClusterCatalog, "giantswarm", "Catalog for cluster app. (OpenStack only).")
	cmd.Flags().StringVar(&f.App.ClusterVersion, flagClusterVersion, "", "Version of cluster to be created. (OpenStack only).")
	cmd.Flags().StringVar(&f.App.DefaultAppsCatalog, flagDefaultAppsCatalog, "giantswarm", "Catalog for cluster default apps app. (OpenStack only).")
	cmd.Flags().StringVar(&f.App.DefaultAppsVersion, flagDefaultAppsVersion, "", "Version of default apps to be created. (OpenStack only).")

	// TODO: Make these flags visible once we have a better method for displaying provider-specific flags.
	_ = cmd.Flags().MarkHidden(flagOpenStackCloud)
	_ = cmd.Flags().MarkHidden(flagOpenStackCloudConfig)
	_ = cmd.Flags().MarkHidden(flagOpenStackDNSNameservers)
	_ = cmd.Flags().MarkHidden(flagOpenStackExternalNetworkID)
	_ = cmd.Flags().MarkHidden(flagOpenStackNodeCIDR)
	_ = cmd.Flags().MarkHidden(flagOpenStackBastionMachineFlavor)
	_ = cmd.Flags().MarkHidden(flagOpenStackBastionDiskSize)
	_ = cmd.Flags().MarkHidden(flagOpenStackBastionImage)
	_ = cmd.Flags().MarkHidden(flagOpenStackControlPlaneDiskSize)
	_ = cmd.Flags().MarkHidden(flagOpenStackControlPlaneImage)
	_ = cmd.Flags().MarkHidden(flagOpenStackControlPlaneMachineFlavor)
	_ = cmd.Flags().MarkHidden(flagOpenStackWorkerDiskSize)
	_ = cmd.Flags().MarkHidden(flagOpenStackWorkerFailureDomain)
	_ = cmd.Flags().MarkHidden(flagOpenStackWorkerImage)
	_ = cmd.Flags().MarkHidden(flagOpenStackWorkerMachineFlavor)
	_ = cmd.Flags().MarkHidden(flagOpenStackWorkerReplicas)

	_ = cmd.Flags().MarkHidden(flagClusterCatalog)
	_ = cmd.Flags().MarkHidden(flagClusterVersion)
	_ = cmd.Flags().MarkHidden(flagDefaultAppsCatalog)
	_ = cmd.Flags().MarkHidden(flagDefaultAppsVersion)

	// Common.
	cmd.Flags().StringSliceVar(&f.ControlPlaneAZ, flagControlPlaneAZ, nil, "Availability zone(s) to use by control plane nodes.")
	cmd.Flags().StringVar(&f.Description, flagDescription, "", "User-friendly description of the cluster's purpose (formerly called name).")
	cmd.Flags().StringVar(&f.KubernetesVersion, flagKubernetesVersion, "v1.20.9", "Cluster Kubernetes version.")
	cmd.Flags().StringVar(&f.Name, flagName, "", "Unique identifier of the cluster (formerly called ID).")
	cmd.Flags().StringVar(&f.OIDC.IssuerURL, flagOIDCIssuerURL, "", "OIDC issuer URL.")
	cmd.Flags().StringVar(&f.OIDC.CAFile, flagOIDCCAFile, "", "OIDC CA FilePath.")
	cmd.Flags().StringVar(&f.OIDC.ClientID, flagOIDCClientID, "", "OIDC client ID.")
	cmd.Flags().StringVar(&f.OIDC.UsernameClaim, flagOIDCUsernameClaim, "", "OIDC username claim.")
	cmd.Flags().StringVar(&f.OIDC.GroupsClaim, flagOIDCGroupsClaim, "", "OIDC groups claim.")
	cmd.Flags().StringVar(&f.Output, flagOutput, "", "File path for storing CRs.")
	cmd.Flags().StringVar(&f.Organization, flagOrganization, "", "Workload cluster organization.")
	cmd.Flags().StringVar(&f.PodsCIDR, flagPodsCIDR, "", "CIDR used for the pods.")
	cmd.Flags().StringVar(&f.Release, flagRelease, "", "Workload cluster release.")
	cmd.Flags().StringSliceVar(&f.Label, flagLabel, nil, "Workload cluster label.")

	// TODO: Make this flag visible when we roll CAPA/EKS out for customers
	_ = cmd.Flags().MarkHidden(flagAWSEKS)

	f.config = genericclioptions.NewConfigFlags(true)
	f.print = genericclioptions.NewPrintFlags("")
	f.print.OutputFormat = nil

	// Merging current command flags and config flags,
	// to be able to override kubectl-specific ones.
	f.config.(*genericclioptions.ConfigFlags).AddFlags(cmd.Flags())
	f.print.AddFlags(cmd)
}

func (f *flag) Validate() error {
	var err error
	validProviders := []string{
		key.ProviderAWS,
		key.ProviderAzure,
		key.ProviderOpenStack,
		key.ProviderVSphere,
	}
	isValidProvider := false
	for _, p := range validProviders {
		if f.Provider == p {
			isValidProvider = true
			break
		}
	}
	if !isValidProvider {
		return microerror.Maskf(invalidFlagError, "--%s must be one of: %s", flagProvider, strings.Join(validProviders, ", "))
	}

	if f.Name != "" {
		if len(f.Name) != key.IDLength {
			return microerror.Maskf(invalidFlagError, "--%s must be of length %d", flagName, key.IDLength)
		}

		matchedLettersOnly, err := regexp.MatchString("^[a-z]+$", f.Name)
		if err == nil && matchedLettersOnly {
			// strings is letters only, which we avoid
			return microerror.Maskf(invalidFlagError, "--%s must contain at least one number", flagName)
		}

		matchedNumbersOnly, err := regexp.MatchString("^[0-9]+$", f.Name)
		if err == nil && matchedNumbersOnly {
			// strings is numbers only, which we avoid
			return microerror.Maskf(invalidFlagError, "--%s must contain at least one letter", flagName)
		}

		matched, err := regexp.MatchString("^[a-z][a-z0-9]+$", f.Name)
		if err == nil && !matched {
			return microerror.Maskf(invalidFlagError, "--%s must only contain alphanumeric characters, and start with a letter", flagName)
		}
	}

	if f.PodsCIDR != "" {
		if !validateCIDR(f.PodsCIDR) {
			return microerror.Maskf(invalidFlagError, "--%s must be a valid CIDR", flagPodsCIDR)
		}
	}

	if f.Organization == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagOrganization)
	}

	{
		// Validate Master AZs.
		switch f.Provider {
		case key.ProviderAWS:
			if len(f.ControlPlaneAZ) != 0 && len(f.ControlPlaneAZ) != 1 && len(f.ControlPlaneAZ) != 3 {
				return microerror.Maskf(invalidFlagError, "--%s must be set to either one or three availability zone names", flagControlPlaneAZ)
			}
			if f.AWS.ControlPlaneSubnet != "" {
				matchedSubnet, err := regexp.MatchString("^20|21|22|23|24|25$", f.AWS.ControlPlaneSubnet)
				if err == nil && !matchedSubnet {
					return microerror.Maskf(invalidFlagError, "--%s must be a valid subnet size (20, 21, 22, 23, 24 or 25)", flagAWSControlPlaneSubnet)
				}
			}
		case key.ProviderAzure:
			if len(f.ControlPlaneAZ) > 1 {
				return microerror.Maskf(invalidFlagError, "--%s supports one availability zone only", flagControlPlaneAZ)
			}
		case key.ProviderOpenStack:
			if f.OpenStack.Cloud == "" {
				return microerror.Maskf(invalidFlagError, "--%s is required", flagOpenStackCloud)
			}
			if f.OpenStack.CloudConfig == "" {
				return microerror.Maskf(invalidFlagError, "--%s is required", flagOpenStackCloudConfig)
			}
			if f.OpenStack.ExternalNetworkID == "" {
				return microerror.Maskf(invalidFlagError, "--%s is required", flagOpenStackExternalNetworkID)
			}
			if f.OpenStack.NodeCIDR != "" {
				if !validateCIDR(f.OpenStack.NodeCIDR) {
					return microerror.Maskf(invalidFlagError, "--%s must be a valid CIDR", flagOpenStackNodeCIDR)
				}
			}
			// bastion
			if f.OpenStack.Bastion.BootFromVolume && f.OpenStack.Bastion.DiskSize < 1 {
				return microerror.Maskf(invalidFlagError, "--%s must be greater than 0 when --%s is specified", flagOpenStackBastionDiskSize, flagOpenStackBastionBootFromVolume)
			}
			if f.OpenStack.Bastion.Flavor == "" {
				return microerror.Maskf(invalidFlagError, "--%s is required", flagOpenStackBastionMachineFlavor)
			}
			if f.OpenStack.Bastion.Image == "" {
				return microerror.Maskf(invalidFlagError, "--%s is required", flagOpenStackBastionImage)
			}
			// control plane
			if f.OpenStack.ControlPlane.BootFromVolume && f.OpenStack.ControlPlane.DiskSize < 1 {
				return microerror.Maskf(invalidFlagError, "--%s must be greater than 0 when --%s is specified", flagOpenStackControlPlaneDiskSize, flagOpenStackControlPlaneBootFromVolume)
			}
			if f.OpenStack.ControlPlane.Flavor == "" {
				return microerror.Maskf(invalidFlagError, "--%s is required", flagOpenStackControlPlaneMachineFlavor)
			}
			if f.OpenStack.ControlPlane.Image == "" {
				return microerror.Maskf(invalidFlagError, "--%s is required", flagOpenStackControlPlaneImage)
			}
			// worker
			if f.OpenStack.WorkerReplicas < 1 {
				return microerror.Maskf(invalidFlagError, "--%s must be greater than 0", flagOpenStackWorkerReplicas)
			}
			if f.OpenStack.WorkerFailureDomain == "" {
				return microerror.Maskf(invalidFlagError, "--%s is required", flagOpenStackWorkerFailureDomain)
			}
			if len(f.ControlPlaneAZ) != 0 {
				if len(f.ControlPlaneAZ)%2 != 1 {
					return microerror.Maskf(invalidFlagError, "--%s must be an odd number number of values (usually 1 or 3 for non-HA and HA respectively)", flagControlPlaneAZ)
				}

				var validFailureDomain bool
				for _, az := range f.ControlPlaneAZ {
					if f.OpenStack.WorkerFailureDomain == az {
						validFailureDomain = true
						break
					}
				}

				if !validFailureDomain {
					return microerror.Maskf(invalidFlagError, "--%s must be among the AZs specified with --%s", flagOpenStackWorkerFailureDomain, flagControlPlaneAZ)
				}
			}
			if f.OpenStack.Worker.BootFromVolume && f.OpenStack.Worker.DiskSize < 1 {
				return microerror.Maskf(invalidFlagError, "--%s must be greater than 0 when --%s is specified", flagOpenStackWorkerDiskSize, flagOpenStackWorkerBootFromVolume)
			}
			if f.OpenStack.Worker.Flavor == "" {
				return microerror.Maskf(invalidFlagError, "--%s is required", flagOpenStackWorkerMachineFlavor)
			}
			if f.OpenStack.Worker.Image == "" {
				return microerror.Maskf(invalidFlagError, "--%s is required", flagOpenStackWorkerImage)
			}
		}
	}

	if f.Provider != "openstack" && f.Release == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagRelease)
	}

	_, err = labels.Parse(f.Label)
	if err != nil {
		return microerror.Maskf(invalidFlagError, "--%s must contain valid label definitions (%s)", flagLabel, err)
	}

	return nil
}

func validateCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)

	return err == nil
}
