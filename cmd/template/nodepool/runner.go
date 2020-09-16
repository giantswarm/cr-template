package nodepool

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/giantswarm/kubectl-gs/pkg/azure"

	"github.com/giantswarm/apiextensions/v2/pkg/id"

	"github.com/giantswarm/kubectl-gs/cmd/template/nodepool/provider"
	"github.com/giantswarm/kubectl-gs/internal/key"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/kubectl-gs/pkg/aws"
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
			PublicSSHKey:                        r.flag.PublicSSHKey,
			Region:                              r.flag.Region,
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
		} else {
			switch r.flag.Provider {
			case key.ProviderAWS:
				config.AvailabilityZones = aws.GetAvailabilityZones(r.flag.NumAvailabilityZones, r.flag.Region)
			case key.ProviderAzure:
				config.AvailabilityZones = azure.GetAvailabilityZones(r.flag.NumAvailabilityZones, r.flag.Region)
			}
		}

		var releaseCollection *release.Release
		{
			c := release.Config{
				Provider: r.flag.Provider,
			}
			releaseCollection, err = release.New(c)
			if err != nil {
				return microerror.Mask(err)
			}
		}
		config.ReleaseComponents = releaseCollection.ReleaseComponents(r.flag.Release)
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
