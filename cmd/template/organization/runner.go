package organization

import (
	"context"
	"fmt"
	"io"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	template "github.com/giantswarm/kubectl-gs/pkg/template/organization"
)

type runner struct {
	flag   *flag
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

	config := template.Config{
		Name: r.flag.Name,
	}

	organizationCR, err := template.NewOrganizationCR(config)
	if err != nil {
		return microerror.Mask(err)
	}

	organizationCRYaml, err := yaml.Marshal(organizationCR)
	if err != nil {
		return microerror.Mask(err)
	}

	fmt.Fprint(r.stdout, string(organizationCRYaml))

	return nil
}
