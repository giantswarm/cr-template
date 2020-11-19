package nodepool

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/giantswarm/apiextensions/v3/pkg/id"

	"github.com/giantswarm/kubectl-gs/cmd/template/nodepool/provider"
	"github.com/giantswarm/kubectl-gs/internal/key"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/kubectl-gs/pkg/release"
)

const (
	nodePoolCRFileName = "nodepoolCR"
)

type runner struct {
	flag   *flag
	logger micrologger.Logger
	stdout io.Writer
	stderr io.Writer
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Set defaults based on provider.
	{
		switch r.flag.Provider {
		case key.ProviderAWS:
			if !cmd.Flags().Changed(flagNodesMin) {
				r.flag.NodesMin = minNodesAWS
			}
			if !cmd.Flags().Changed(flagNodesMax) {
				r.flag.NodesMax = maxNodesAWS
			}
		case key.ProviderAzure:
			if !cmd.Flags().Changed(flagNodesMin) {
				r.flag.NodesMin = minNodesAzure
			}
			if !cmd.Flags().Changed(flagNodesMax) {
				r.flag.NodesMax = maxNodesAzure
			}
		}
	}

	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.run(ctx, cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	var err error

	var config provider.NodePoolCRsConfig
	{
		config = provider.NodePoolCRsConfig{
			AWSInstanceType:                     r.flag.AWSInstanceType,
			FileName:                            nodePoolCRFileName,
			ClusterID:                           r.flag.ClusterID,
			Description:                         r.flag.NodepoolName,
			VMSize:                              r.flag.AzureVMSize,
			NodesMax:                            r.flag.NodesMax,
			NodesMin:                            r.flag.NodesMin,
			OnDemandBaseCapacity:                r.flag.OnDemandBaseCapacity,
			OnDemandPercentageAboveBaseCapacity: r.flag.OnDemandPercentageAboveBaseCapacity,
			Owner:                               r.flag.Owner,
			UseAlikeInstanceTypes:               r.flag.UseAlikeInstanceTypes,
			ReleaseVersion:                      r.flag.Release,
		}

		if config.NodePoolID == "" {
			config.NodePoolID = id.Generate()
		}

		// Remove leading 'v' from release flag input.
		config.ReleaseVersion = strings.TrimLeft(config.ReleaseVersion, "v")

		if len(r.flag.AvailabilityZones) > 0 {
			config.AvailabilityZones = r.flag.AvailabilityZones
		}

		var releaseCollection *release.Release
		{
			c := release.Config{
				Provider: r.flag.Provider,
				Branch:   r.flag.ReleaseBranch,
			}
			releaseCollection, err = release.New(c)
			if err != nil {
				return microerror.Mask(err)
			}
		}
		config.ReleaseComponents = releaseCollection.ReleaseComponents(r.flag.Release)

		if r.flag.Provider == key.ProviderAzure {
			config.Namespace = key.OrganizationNamespaceFromName(config.Owner)
		}
	}

	var output *os.File
	{
		if r.flag.Output == "" {
			output = os.Stdout
		} else {
			f, err := os.Create(r.flag.Output)
			if err != nil {
				return microerror.Mask(err)
			}
			defer f.Close()

			output = f
		}
	}

	switch r.flag.Provider {
	case key.ProviderAWS:
		err = provider.WriteAWSTemplate(output, config)
		if err != nil {
			return microerror.Mask(err)
		}
	case key.ProviderAzure:
		err = provider.WriteAzureTemplate(output, config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
