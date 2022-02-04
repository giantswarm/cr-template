package login

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/giantswarm/kubectl-gs/pkg/installation"
	"github.com/giantswarm/kubectl-gs/pkg/kubeconfig"
	"github.com/giantswarm/kubectl-gs/pkg/oidc"
)

type authInfo struct {
	username string
	token    string

	// OIDC-specific.
	clientID     string
	email        string
	refreshToken string
}

// storeMCCredentials stores the installation's CA certificate, and
// updates the kubeconfig with the configuration for the k8s api access.
func storeMCCredentials(k8sConfigAccess clientcmd.ConfigAccess, i *installation.Installation, authResult authInfo, fs afero.Fs, internalAPI bool) error {
	config, err := k8sConfigAccess.GetStartingConfig()
	if err != nil {
		return microerror.Mask(err)
	}

	kUsername := fmt.Sprintf("gs-%s-%s", authResult.username, i.Codename)
	contextName := kubeconfig.GenerateKubeContextName(i.Codename)
	clusterName := fmt.Sprintf("gs-%s", i.Codename)

	// Store CA certificate.
	err = kubeconfig.WriteCertificate(i.CACert, clusterName, fs)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		// Create authenticated user.
		initialUser, exists := config.AuthInfos[kUsername]
		if !exists {
			initialUser = clientcmdapi.NewAuthInfo()
		}

		if len(authResult.clientID) > 0 {
			initialUser.AuthProvider = &clientcmdapi.AuthProviderConfig{
				Name: "oidc",
				Config: map[string]string{
					ClientID:     authResult.clientID,
					IDToken:      authResult.token,
					Issuer:       i.AuthURL,
					RefreshToken: authResult.refreshToken,
				},
			}
		} else {
			initialUser.Token = authResult.token
		}

		// Add user information to config.
		config.AuthInfos[kUsername] = initialUser
	}

	{
		// Create authenticated cluster.
		initialCluster, exists := config.Clusters[clusterName]
		if !exists {
			initialCluster = clientcmdapi.NewCluster()
		}

		if internalAPI {
			initialCluster.Server = i.K8sInternalApiURL
		} else {
			initialCluster.Server = i.K8sApiURL
		}

		var certPath string
		certPath, err = kubeconfig.GetKubeCertFilePath(clusterName)
		if err != nil {
			return microerror.Mask(err)
		}
		initialCluster.CertificateAuthority = certPath

		// Add cluster configuration to config.
		config.Clusters[clusterName] = initialCluster
	}

	{
		// Create authenticated context.
		initialContext, exists := config.Contexts[contextName]
		if !exists {
			initialContext = clientcmdapi.NewContext()
		}

		initialContext.Cluster = clusterName

		initialContext.AuthInfo = kUsername

		// Add context configuration to config.
		config.Contexts[contextName] = initialContext

		// Select newly created context as current.
		config.CurrentContext = contextName
	}

	err = clientcmd.ModifyConfig(k8sConfigAccess, *config, false)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// printMCCredentials saves the installation's CA certificate, and
// writes the configuration for the k8s api access into a separate file.
func printMCCredentials(k8sConfigAccess clientcmd.ConfigAccess, i *installation.Installation, authResult authInfo, fs afero.Fs, internalAPI bool, filePath string) error {
	kUsername := fmt.Sprintf("gs-%s-%s", authResult.username, i.Codename)
	contextName := kubeconfig.GenerateKubeContextName(i.Codename)
	clusterName := fmt.Sprintf("gs-%s", i.Codename)

	// Store CA certificate.
	err := kubeconfig.WriteCertificate(i.CACert, clusterName, fs)
	if err != nil {
		return microerror.Mask(err)
	}

	var server string
	{
		if internalAPI {
			server = i.K8sInternalApiURL
		} else {
			server = i.K8sApiURL
		}
	}

	var certPath string
	{
		certPath, err = kubeconfig.GetKubeCertFilePath(clusterName)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	authInfo := clientcmdapi.NewAuthInfo()
	{
		if len(authResult.clientID) > 0 {
			authInfo.AuthProvider = &clientcmdapi.AuthProviderConfig{
				Name: "oidc",
				Config: map[string]string{
					ClientID:     authResult.clientID,
					IDToken:      authResult.token,
					Issuer:       i.AuthURL,
					RefreshToken: authResult.refreshToken,
				},
			}
		} else {
			authInfo.Token = authResult.token
		}
	}

	kubeconfig := clientcmdapi.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: map[string]*clientcmdapi.Cluster{
			contextName: {
				Server:               server,
				CertificateAuthority: certPath,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			contextName: {
				Cluster:  clusterName,
				AuthInfo: kUsername,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			kUsername: authInfo,
		},
		CurrentContext: contextName,
	}
	if exists, err := afero.Exists(fs, filePath); exists {
		return microerror.Maskf(fileExistsError, "The destination file %s already exists. Please specify a different destination.", filePath)
	} else if err != nil {
		return microerror.Mask(err)
	}
	err = clientcmd.WriteToFile(kubeconfig, filePath)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

// switchContext modifies the existing kubeconfig, and switches the currently
// active context to the one specified.
func switchContext(ctx context.Context, k8sConfigAccess clientcmd.ConfigAccess, newContextName string) error {
	config, err := k8sConfigAccess.GetStartingConfig()
	if err != nil {
		return microerror.Mask(err)
	}

	// Check if the context exists.
	if _, exists := config.Contexts[newContextName]; !exists {
		return microerror.Maskf(contextDoesNotExistError, "There is no context named '%s'. Please make sure you spelled the installation handle correctly.\nIf not sure, pass the Management API URL or the web UI URL of the installation as an argument.", newContextName)
	}

	authType := kubeconfig.GetAuthType(config, newContextName)
	if authType == kubeconfig.AuthTypeAuthProvider {
		authProvider, exists := kubeconfig.GetAuthProvider(config, newContextName)
		if !exists {
			return microerror.Maskf(incorrectConfigurationError, "There is no authentication configuration for the '%s' context", newContextName)
		}

		err = validateOIDCProvider(authProvider)
		if IsNewLoginRequired(err) {
			return microerror.Mask(err)
		} else if err != nil {
			return microerror.Maskf(incorrectConfigurationError, "The authentication configuration is corrupted, please log in again using a URL.")
		}

		if newContextName == config.CurrentContext {
			return microerror.Mask(contextAlreadySelectedError)
		}

		var auther *oidc.Authenticator
		{
			oidcConfig := oidc.Config{
				Issuer:   authProvider.Config[Issuer],
				ClientID: authProvider.Config[ClientID],
			}

			auther, err = oidc.New(ctx, oidcConfig)
			if err != nil {
				return microerror.Maskf(incorrectConfigurationError, "\n%v", err.Error())
			}
		}

		// Renew authentication token.
		{
			idToken, rToken, err := auther.RenewToken(ctx, authProvider.Config[RefreshToken])
			if err != nil {
				return microerror.Mask(tokenRenewalFailedError)
			}
			authProvider.Config[RefreshToken] = rToken
			authProvider.Config[IDToken] = idToken
		}
	} else if authType == kubeconfig.AuthTypeUnknown {
		return microerror.Maskf(incorrectConfigurationError, "There is no authentication configuration for the '%s' context", newContextName)
	}

	config.CurrentContext = newContextName

	err = clientcmd.ModifyConfig(k8sConfigAccess, *config, true)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
