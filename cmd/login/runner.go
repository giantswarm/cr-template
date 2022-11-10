package login

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/giantswarm/kubectl-gs/v2/pkg/commonconfig"
	"github.com/giantswarm/kubectl-gs/v2/pkg/kubeconfig"
)

type runner struct {
	flag   *flag
	logger micrologger.Logger
	fs     afero.Fs

	commonConfig *commonconfig.CommonConfig
	loginOptions LoginOptions

	stdout io.Writer
	stderr io.Writer
}

type LoginOptions struct {
	isWCClientCert            bool
	selfContained             bool
	selfContainedClientCert   bool
	switchToContext           bool
	switchToClientCertContext bool
	originContext             string
	contextOverride           string
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	r.setLoginOptions(ctx, &args)
	err = r.run(ctx, cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	switch len(args) {
	// No arguments given - we try to reuse the existing context.
	case 0:
		err := r.tryToReuseExistingContext(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
	// One argument - This can be an existing context, an installation name or a login URL
	case 1:
		installationIdentifier := strings.ToLower(args[0])
		foundContext, err := r.findContext(ctx, installationIdentifier)
		if IsContextDoesNotExist(err) && !strings.HasSuffix(installationIdentifier, kubeconfig.ClientCertSuffix) {
			clientCertContext := kubeconfig.GetClientCertContextName(installationIdentifier)
			fmt.Fprint(r.stdout, color.YellowString("No context named %s was found: %s\nLooking for context %s.\n", installationIdentifier, err, clientCertContext))
			foundContext, err = r.findContext(ctx, clientCertContext)
		}
		if err != nil {
			return microerror.Mask(err)
		}
		if !foundContext {
			err = r.loginWithURL(ctx, installationIdentifier, true, r.commonConfig.GetTokenOverride())
			if err != nil {
				return microerror.Mask(err)
			}
		}
	// Two arguments - This is interpreted as an installation and a workload cluster
	case 2:
		installationIdentifier := strings.ToLower(strings.Join(args, "-"))
		foundContext, err := r.findContext(ctx, installationIdentifier)
		if IsContextDoesNotExist(err) && !strings.HasSuffix(installationIdentifier, kubeconfig.ClientCertSuffix) {
			clientCertContext := kubeconfig.GetClientCertContextName(installationIdentifier)
			fmt.Fprint(r.stdout, color.YellowString("No context named %s was found: %s\nLooking for context %s.\n", installationIdentifier, err, clientCertContext))
			foundContext, err = r.findContext(ctx, clientCertContext)
		}
		if err != nil {
			return microerror.Mask(err)
		}
		if !foundContext {
			return microerror.Maskf(contextDoesNotExistError, "Could not find context for identifier %s", installationIdentifier)
		}
	default:
		return microerror.Maskf(invalidConfigError, "Invalid number of arguments.")
	}

	// Clientcert creation if desired
	if r.loginOptions.isWCClientCert {
		return r.handleWCClientCert(ctx)
	}

	return nil
}

func (r *runner) tryToGetCurrentContexts(ctx context.Context) (string, string, error) {
	config, err := r.commonConfig.GetConfigAccess().GetStartingConfig()
	if err != nil {
		return "", "", microerror.Mask(err)
	}
	contextOverride := r.commonConfig.GetContextOverride()
	if _, ok := config.Contexts[contextOverride]; !ok {
		contextOverride = ""
	}
	return config.CurrentContext, contextOverride, nil
}

func (r *runner) setLoginOptions(ctx context.Context, args *[]string) {
	originContext, contextOverride, err := r.tryToGetCurrentContexts(ctx)
	if err != nil {
		fmt.Fprintln(r.stdout, color.YellowString("Failed trying to determine current context. %s", err))
	}
	r.loginOptions = LoginOptions{
		originContext:           originContext,
		contextOverride:         contextOverride,
		isWCClientCert:          len(r.flag.WCName) > 0,
		selfContained:           len(r.flag.SelfContained) > 0 && !(len(r.flag.WCName) > 0),
		selfContainedClientCert: len(r.flag.SelfContained) > 0 && len(r.flag.WCName) > 0,
	}
	r.loginOptions.switchToContext = r.loginOptions.contextOverride == "" && (r.loginOptions.isWCClientCert || !(r.loginOptions.selfContained || r.flag.KeepContext))
	r.loginOptions.switchToClientCertContext = r.loginOptions.isWCClientCert && !(r.loginOptions.selfContainedClientCert || r.flag.KeepContext)
}

func (r *runner) tryToReuseExistingContext(ctx context.Context) error {
	currentContext := r.commonConfig.GetContextOverride()
	if currentContext == "" {
		config, err := r.commonConfig.GetConfigAccess().GetStartingConfig()
		if err != nil {
			return microerror.Mask(err)
		}
		currentContext = config.CurrentContext
	}
	if currentContext != "" {
		return r.loginWithKubeContextName(ctx, currentContext)
	} else {
		return microerror.Maskf(selectedContextNonCompatibleError, "The current context does not seem to belong to a Giant Swarm cluster.\nPlease run 'kubectl gs login --help' to find out how to log in to a particular cluster.")
	}
}
