package nodepool

import (
	"context"
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/kubectl-gs/internal/key"
	"github.com/giantswarm/kubectl-gs/pkg/aws"
	"github.com/giantswarm/kubectl-gs/pkg/gsrelease"
	"github.com/giantswarm/kubectl-gs/pkg/template/nodepool"
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

	var availabilityZones []string
	{
		if r.flag.AvailabilityZones != "" {
			availabilityZones = strings.Split(r.flag.AvailabilityZones, ",")
		} else {
			availabilityZones = aws.GetAvailabilityZones(r.flag.NumAvailabilityZones, r.flag.Region)
		}
	}

	var release *gsrelease.GSRelease
	{
		c := gsrelease.Config{
			NoCache: r.flag.NoCache,
		}

		release, err = gsrelease.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	releaseComponents := release.ReleaseComponents(r.flag.Release)

	config := nodepool.Config{
		AvailabilityZones:                   availabilityZones,
		AWSInstanceType:                     r.flag.AWSInstanceType,
		ClusterID:                           r.flag.ClusterID,
		Name:                                r.flag.NodepoolName,
		NodesMax:                            r.flag.NodesMax,
		NodesMin:                            r.flag.NodesMin,
		OnDemandBaseCapacity:                r.flag.OnDemandBaseCapacity,
		OnDemandPercentageAboveBaseCapacity: r.flag.OnDemandPercentageAboveBaseCapacity,
		Owner:                               r.flag.Owner,
		ReleaseComponents:                   releaseComponents,
		ReleaseVersion:                      r.flag.Release,
	}

	mdCR, awsMDCR, err := nodepool.NewMachineDeploymentCRs(config)
	if err != nil {
		return microerror.Mask(err)
	}

	mdCRYaml, err := yaml.Marshal(mdCR)
	if err != nil {
		return microerror.Mask(err)
	}

	awsMDCRYaml, err := yaml.Marshal(awsMDCR)
	if err != nil {
		return microerror.Mask(err)
	}

	type MachineDeploymentCRsOutput struct {
		AWSMachineDeploymentCR string
		MachineDeploymentCR    string
	}

	nodepoolCRsOutput := MachineDeploymentCRsOutput{
		AWSMachineDeploymentCR: string(awsMDCRYaml),
		MachineDeploymentCR:    string(mdCRYaml),
	}

	t := template.Must(template.New("nodepoolCR").Parse(key.MachineDeploymentCRsTemplate))

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

		err = t.Execute(output, nodepoolCRsOutput)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
