package renewtoken

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/giantswarm/kubectl-gs/pkg/kubeconfig"
	"github.com/giantswarm/kubectl-gs/pkg/middleware"
	"github.com/giantswarm/kubectl-gs/pkg/oidc"
)

// Middleware will attempt to renew the current context's auth info token.
// If the renewal fails, this middleware will not fail.
func Middleware(k8sConfigAccess clientcmd.ConfigAccess) middleware.Middleware {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		config, err := k8sConfigAccess.GetStartingConfig()
		if err != nil {
			return nil
		}

		authProvider, exists := kubeconfig.GetAuthProvider(config, config.CurrentContext)
		if !exists {
			return nil
		}

		var auther *oidc.Authenticator
		{
			oidcConfig := oidc.Config{
				Issuer:       authProvider.Config["idp-issuer-url"],
				ClientID:     authProvider.Config["client-id"],
				ClientSecret: authProvider.Config["client-secret"],
			}
			auther, err = oidc.New(ctx, oidcConfig)
			if err != nil {
				return nil
			}
		}

		{
			idToken, rToken, err := auther.RenewToken(ctx, authProvider.Config["refresh-token"])
			if err != nil {
				return nil
			}
			authProvider.Config["refresh-token"] = rToken
			authProvider.Config["id-token"] = idToken
		}

		_ = clientcmd.ModifyConfig(k8sConfigAccess, *config, true)

		return nil
	}
}
