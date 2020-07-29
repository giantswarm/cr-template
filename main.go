package main

import (
	"context"
	"fmt"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/giantswarm/kubectl-gs/cmd"
	"github.com/giantswarm/kubectl-gs/pkg/errorprinter"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func main() {
	err := mainE(context.Background())
	if err != nil {
		if isDebugMode() {
			panic(microerror.JSON(err))
		} else {
			ep := errorprinter.New()
			fmt.Print(ep.Format(err))
			os.Exit(1)
		}
	}
}

func mainE(ctx context.Context) error {
	var err error

	var logger micrologger.Logger
	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	fs := afero.NewOsFs()

	k8sConfigAccess := clientcmd.NewDefaultPathOptions()

	var rootCommand *cobra.Command
	{
		c := cmd.Config{
			Logger:     logger,
			FileSystem: fs,

			K8sConfigAccess: k8sConfigAccess,
		}

		rootCommand, err = cmd.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = rootCommand.Execute()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func isDebugMode() bool {
	for _, arg := range os.Args {
		if arg == "--debug" {
			return true
		}
	}

	return false
}
