package cluster

import (
	"net"
	"regexp"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/kubectl-gs/internal/key"
	"github.com/giantswarm/kubectl-gs/pkg/clusterlabels"
)

const (
	flagProvider = "provider"

	// AWS only.
	flagExternalSNAT       = "external-snat"
	flagPodsCIDR           = "pods-cidr"
	flagControlPlaneSubnet = "control-plane-subnet"

	// Common.
	flagClusterID = "cluster-id"
	flagMasterAZ  = "master-az"
	flagName      = "name"
	flagOutput    = "output"
	flagOwner     = "owner"
	flagRelease   = "release"
	flagLabel     = "label"
)

type flag struct {
	Provider string

	// AWS only.
	ControlPlaneSubnet string
	ExternalSNAT       bool
	PodsCIDR           string

	// Common.
	ClusterID string
	MasterAZ  []string
	Name      string
	Output    string
	Owner     string
	Release   string
	Label     []string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.Provider, flagProvider, "", "Installation infrastructure provider.")

	// AWS only.
	cmd.Flags().StringVar(&f.ControlPlaneSubnet, flagControlPlaneSubnet, "", "Subnet used for the Control Plane.")
	cmd.Flags().BoolVar(&f.ExternalSNAT, flagExternalSNAT, false, "AWS CNI configuration.")
	cmd.Flags().StringVar(&f.PodsCIDR, flagPodsCIDR, "", "CIDR used for the pods.")

	// Common.
	cmd.Flags().StringVar(&f.ClusterID, flagClusterID, "", "User-defined cluster ID.")
	cmd.Flags().StringSliceVar(&f.MasterAZ, flagMasterAZ, []string{}, "Workload master node availability zone.")
	cmd.Flags().StringVar(&f.Name, flagName, "", "Workload cluster name.")
	cmd.Flags().StringVar(&f.Output, flagOutput, "", "File path for storing CRs.")
	cmd.Flags().StringVar(&f.Owner, flagOwner, "", "Workload cluster owner organization.")
	cmd.Flags().StringVar(&f.Release, flagRelease, "", "Workload cluster release. If not given, this remains empty for defaulting to the most recent one via the Management API.")
	cmd.Flags().StringSliceVar(&f.Label, flagLabel, nil, "Workload cluster label.")
}

func (f *flag) Validate() error {
	var err error

	if f.Provider != key.ProviderAWS && f.Provider != key.ProviderAzure {
		return microerror.Maskf(invalidFlagError, "--%s must be either aws or azure", flagProvider)
	}

	if f.ClusterID != "" {
		if len(f.ClusterID) != key.IDLength {
			return microerror.Maskf(invalidFlagError, "--%s must be length of %d", flagClusterID, key.IDLength)
		}

		matchedLettersOnly, err := regexp.MatchString("^[a-z]+$", f.ClusterID)
		if err == nil && matchedLettersOnly {
			// strings is letters only, which we avoid
			return microerror.Maskf(invalidFlagError, "--%s must contain at least one number", flagClusterID)
		}

		matchedNumbersOnly, err := regexp.MatchString("^[0-9]+$", f.ClusterID)
		if err == nil && matchedNumbersOnly {
			// strings is numbers only, which we avoid
			return microerror.Maskf(invalidFlagError, "--%s must contain at least one digit", flagClusterID)
		}

		matched, err := regexp.MatchString("^[a-z][a-z0-9]+$", f.ClusterID)
		if err == nil && !matched {
			return microerror.Maskf(invalidFlagError, "--%s must only contain alphanumeric characters, and start with a letter", flagClusterID)
		}

		if f.ControlPlaneSubnet != "" {
			matchedSubnet, err := regexp.MatchString("^24|25|26|27$", f.ControlPlaneSubnet)
			if err == nil && !matchedSubnet {
				return microerror.Maskf(invalidFlagError, "--%s must be a valid subnet size (24, 25, 26 or 27)", flagControlPlaneSubnet)
			}
		}

		return nil
	}
	// Validate name for non-aws clusters.
	if f.Provider != key.ProviderAWS && f.Name == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagName)
	}
	if f.PodsCIDR != "" {
		if !validateCIDR(f.PodsCIDR) {
			return microerror.Maskf(invalidFlagError, "--%s must be a valid CIDR", flagPodsCIDR)
		}
	}
	if f.Owner == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagOwner)
	}

	{
		// Validate Master AZs.
		switch f.Provider {
		case key.ProviderAWS:
			if len(f.MasterAZ) != 0 && len(f.MasterAZ) != 1 && len(f.MasterAZ) != 3 {
				return microerror.Maskf(invalidFlagError, "--%s must be set to either one or three availability zone names", flagMasterAZ)
			}
		case key.ProviderAzure:
			if len(f.MasterAZ) > 1 {
				return microerror.Maskf(invalidFlagError, "--%s supports one availability zone only", flagMasterAZ)
			}
		}
	}

	// Validate release version for non-aws clusters.
	if f.Provider != key.ProviderAWS && f.Release == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagRelease)
	}

	_, err = clusterlabels.Parse(f.Label)
	if err != nil {
		return microerror.Maskf(invalidFlagError, "--%s must contain valid label definitions (%s)", flagLabel, err)
	}

	return nil
}

func validateCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)

	return err == nil
}
