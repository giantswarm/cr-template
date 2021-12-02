package login

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/giantswarm/kubectl-gs/internal/feature"
	"github.com/giantswarm/kubectl-gs/internal/key"
	kgslabel "github.com/giantswarm/kubectl-gs/internal/label"
	"github.com/giantswarm/kubectl-gs/pkg/data/domain/clientcert"
	"github.com/giantswarm/kubectl-gs/pkg/data/domain/cluster"
	"github.com/giantswarm/kubectl-gs/pkg/data/domain/organization"
	"github.com/giantswarm/kubectl-gs/pkg/data/domain/release"
	"github.com/giantswarm/kubectl-gs/pkg/kubeconfig"
)

const (
	credentialRetryTimeout    = 1 * time.Second
	credentialMaxRetryTimeout = 5 * time.Minute

	credentialKeyCertCRT = "crt"
	credentialKeyCertKey = "key"
	credentialKeyCertCA  = "ca"
)

type clientCertConfig struct {
	provider            string
	clusterName         string
	clusterNamespace    string
	organizationName    string
	ttl                 string
	groups              []string
	clusterBasePath     string
	certOperatorVersion string
}

func generateClientCertUID() string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(time.Now().String()))

	uid := fmt.Sprintf("%x", hash.Sum(nil))

	return uid[:16]
}

func generateClientCert(config clientCertConfig) (*clientcert.ClientCert, error) {
	clientCertUID := generateClientCertUID()
	clientCertName := fmt.Sprintf("%s-%s", config.clusterName, clientCertUID)
	commonName := fmt.Sprintf("%s.%s.k8s.%s", clientCertUID, config.clusterName, config.clusterBasePath)

	certConfig := &corev1alpha1.CertConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clientCertName,
			Namespace: config.clusterNamespace,
			Labels: map[string]string{
				kgslabel.CertOperatorVersion: config.certOperatorVersion,
				kgslabel.Certificate:         clientCertUID,
				label.Cluster:                config.clusterName,
				label.Organization:           config.organizationName,
			},
		},
		Spec: corev1alpha1.CertConfigSpec{
			Cert: corev1alpha1.CertConfigSpecCert{
				AllowBareDomains:    true,
				ClusterComponent:    clientCertUID,
				ClusterID:           config.clusterName,
				CommonName:          commonName,
				DisableRegeneration: true,
				Organizations:       config.groups,
				TTL:                 config.ttl,
			},
			VersionBundle: corev1alpha1.CertConfigSpecVersionBundle{
				Version: config.certOperatorVersion,
			},
		},
	}

	r := &clientcert.ClientCert{
		CertConfig: certConfig,
	}

	return r, nil
}

func createCert(ctx context.Context, clientCertService clientcert.Interface, config clientCertConfig) (*clientcert.ClientCert, error) {
	clientCert, err := generateClientCert(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = clientCertService.Create(ctx, clientCert)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return clientCert, nil
}

// fetchCredential tries to fetch the client certificate credential
// for a couple of times, until the resource is fetched, or until the timeout is reached.
func fetchCredential(ctx context.Context, provider string, clientCertService clientcert.Interface, clientCert *clientcert.ClientCert) (*corev1.Secret, error) {
	credentialNamespace := determineCredentialNamespace(clientCert, provider)

	var secret *corev1.Secret
	var err error

	o := func() error {
		secret, err = clientCertService.GetCredential(ctx, credentialNamespace, clientCert.CertConfig.Name)
		if clientcert.IsNotFound(err) {
			// Client certificate credential has not been created yet, try again until it is.
			return microerror.Mask(err)
		} else if err != nil {
			return backoff.Permanent(microerror.Mask(err))
		}

		return nil
	}
	b := backoff.NewConstant(credentialMaxRetryTimeout, credentialRetryTimeout)

	err = backoff.Retry(o, b)
	if clientcert.IsNotFound(err) {
		return nil, microerror.Maskf(credentialRetrievalTimedOut, "failed to get the client certificate credential on time")
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	return secret, nil
}

// storeCredential saves the created client certificate credentials into the kubectl config.
func storeCredential(k8sConfigAccess clientcmd.ConfigAccess, fs afero.Fs, clientCert *clientcert.ClientCert, credential *corev1.Secret, clusterBasePath string) (string, bool, error) {
	config, err := k8sConfigAccess.GetStartingConfig()
	if err != nil {
		return "", false, microerror.Mask(err)
	}

	mcContextName := config.CurrentContext
	contextName := kubeconfig.GenerateWCKubeContextName(mcContextName, clientCert.CertConfig.Spec.Cert.ClusterID)
	userName := fmt.Sprintf("%s-user", contextName)
	clusterName := contextName
	clusterServer := fmt.Sprintf("https://api.%s.k8s.%s", clientCert.CertConfig.Spec.Cert.ClusterID, clusterBasePath)

	certCRT := credential.Data[credentialKeyCertCRT]
	certKey := credential.Data[credentialKeyCertKey]
	certCA := credential.Data[credentialKeyCertCA]

	contextExists := false

	{
		// Create authenticated user.
		user, exists := config.AuthInfos[userName]
		if !exists {
			user = clientcmdapi.NewAuthInfo()
		}

		user.ClientCertificateData = certCRT
		user.ClientKeyData = certKey

		// Add user information to config.
		config.AuthInfos[userName] = user
	}

	{
		// Create authenticated cluster.
		cluster, exists := config.Clusters[clusterName]
		if !exists {
			cluster = clientcmdapi.NewCluster()
		}

		cluster.Server = clusterServer
		cluster.CertificateAuthorityData = certCA

		// Add cluster configuration to config.
		config.Clusters[clusterName] = cluster
	}

	{
		// Create authenticated context.
		var context *clientcmdapi.Context
		context, contextExists = config.Contexts[contextName]
		if !contextExists {
			context = clientcmdapi.NewContext()
		}

		context.Cluster = clusterName
		context.AuthInfo = userName

		// Add context configuration to config.
		config.Contexts[contextName] = context

		// Select newly created context as current.
		config.CurrentContext = contextName
	}

	err = clientcmd.ModifyConfig(k8sConfigAccess, *config, false)
	if err != nil {
		return "", contextExists, microerror.Mask(err)
	}

	return contextName, contextExists, nil
}

// printCredential saves the created client certificate credentials into a separate kubectl config file.
func printCredential(k8sConfigAccess clientcmd.ConfigAccess, fs afero.Fs, filePath string, clientCert *clientcert.ClientCert, credential *corev1.Secret, clusterBasePath string) (string, bool, error) {
	config, err := k8sConfigAccess.GetStartingConfig()
	if err != nil {
		return "", false, microerror.Mask(err)
	}

	mcContextName := config.CurrentContext
	contextName := kubeconfig.GenerateWCKubeContextName(mcContextName, clientCert.CertConfig.Spec.Cert.ClusterID)

	kubeconfig := clientcmdapi.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: map[string]*clientcmdapi.Cluster{
			contextName: {
				Server:                   fmt.Sprintf("https://api.%s.k8s.%s", clientCert.CertConfig.Spec.Cert.ClusterID, clusterBasePath),
				CertificateAuthorityData: credential.Data[credentialKeyCertCA],
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			contextName: {
				Cluster:  contextName,
				AuthInfo: fmt.Sprintf("%s-user", contextName),
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			fmt.Sprintf("%s-user", contextName): {
				ClientCertificateData: credential.Data[credentialKeyCertCRT],
				ClientKeyData:         credential.Data[credentialKeyCertKey],
			},
		},
		CurrentContext: contextName,
	}
	if exists, err := afero.Exists(fs, filePath); exists {
		return "", false, microerror.Maskf(fileExistsError, "The destination file %s already exists. Please specify a different destination.", filePath)
	} else if err != nil {
		return "", false, microerror.Mask(err)
	}
	err = clientcmd.WriteToFile(kubeconfig, filePath)
	if err != nil {
		return "", false, microerror.Mask(err)
	}
	return contextName, false, nil
}

func cleanUpClientCertResources(ctx context.Context, clientCertService clientcert.Interface, clientCertResource *clientcert.ClientCert) error {
	err := clientCertService.Delete(ctx, clientCertResource)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func getOrganizationNamespace(ctx context.Context, organizationService organization.Interface, name string) (string, error) {
	org, err := organizationService.Get(ctx, organization.GetOptions{Name: name})
	if organization.IsNotFound(err) {
		return "", microerror.Maskf(organizationNotFoundError, "The organization %s could not be found.", name)
	} else if err != nil {
		return "", microerror.Mask(err)
	}

	namespace := org.(*organization.Organization).Organization.Status.Namespace
	if len(namespace) < 1 {
		return "", microerror.Maskf(unknownOrganizationNamespaceError, "Could not find the namespace for organization %s.", name)
	}

	return namespace, nil
}

func getAllOrganizationNamespaces(ctx context.Context, organizationService organization.Interface) ([]string, error) {
	orgList, err := organizationService.Get(ctx, organization.GetOptions{})
	if organization.IsNoResources(err) {
		return nil, microerror.Maskf(noOrganizationsError, "Could not find any organizations.")
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	var namespaces []string
	for _, org := range orgList.(*organization.Collection).Items {
		if len(org.Organization.Status.Namespace) > 0 {
			namespaces = append(namespaces, org.Organization.Status.Namespace)
		}
	}

	return namespaces, nil
}

func determineCredentialNamespace(clientCert *clientcert.ClientCert, provider string) string {
	if provider == key.ProviderAzure {
		return metav1.NamespaceDefault
	}

	return clientCert.CertConfig.GetNamespace()
}

func fetchCluster(ctx context.Context, clusterService cluster.Interface, provider, namespace, name string) (*cluster.Cluster, error) {
	o := cluster.GetOptions{
		Namespace: namespace,
		Name:      name,
		Provider:  provider,
	}

	c, err := clusterService.Get(ctx, o)
	if cluster.IsNotFound(err) {
		return nil, microerror.Maskf(clusterNotFoundError, "The workload cluster %s could not be found.", name)
	} else if cluster.IsInsufficientPermissions(err) {
		return nil, microerror.Maskf(insufficientPermissionsError, "You don't have the required permissions to get clusters in the %s namespace.", namespace)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	return c.(*cluster.Cluster), nil
}

func findCluster(ctx context.Context, clusterService cluster.Interface, organizationService organization.Interface, provider, name string, namespaces ...string) (*cluster.Cluster, error) {
	if len(namespaces) == 1 {
		c, err := fetchCluster(ctx, clusterService, provider, namespaces[0], name)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		return c, nil
	}

	clustersCh := make(chan *cluster.Cluster, len(namespaces))
	errorsCh := make(chan error, len(namespaces))

	var wg sync.WaitGroup

	for _, namespace := range namespaces {
		wg.Add(1)

		go func(namespace string) {
			defer wg.Done()

			c, err := fetchCluster(ctx, clusterService, provider, namespace, name)
			if IsClusterNotFound(err) || IsInsufficientPermissions(err) {
				// Nothing to see here.
				return
			} else if err != nil {
				errorsCh <- microerror.Mask(err)
				return
			}

			clustersCh <- c
		}(namespace)
	}

	wg.Wait()

	close(clustersCh)
	close(errorsCh)

	if len(errorsCh) > 0 {
		return nil, microerror.Mask(<-errorsCh)
	}

	switch len(clustersCh) {
	case 1:
		return <-clustersCh, nil

	case 0:
		return nil, microerror.Maskf(clusterNotFoundError, "The workload cluster %s could not be found.\nMake sure you have access to the cluster's organization namespace.", name)

	default:
		{
			var errMsg strings.Builder
			errMsg.WriteString(fmt.Sprintf("There are multiple workload clusters with the name %s:\n", name))

			i := 1
			for c := range clustersCh {
				org := c.Cluster.Labels[label.Organization]
				if len(org) < 1 {
					org = "n/a"
				}

				errMsg.WriteString(fmt.Sprintf("%d. %s in organization %s\n", i, name, org))

				i++
			}

			errMsg.WriteString(fmt.Sprintf("\nUse the --%s flag to select one from a specific organization.", flagWCOrganization))

			return nil, microerror.Maskf(multipleClustersFoundError, errMsg.String())
		}
	}
}

func getClusterReleaseVersion(c *cluster.Cluster, provider string) (string, error) {
	if c.Cluster.Labels == nil {
		return "", microerror.Maskf(invalidReleaseVersionError, "The workload cluster %s does not have a release version label.", name)
	}

	releaseVersion := c.Cluster.Labels[label.ReleaseVersion]
	if len(releaseVersion) < 1 {
		return "", microerror.Maskf(invalidReleaseVersionError, "The workload cluster %s has an invalid release version.", name)
	}

	err := validateReleaseVersion(releaseVersion, provider)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return releaseVersion, nil
}

func getCertOperatorVersion(ctx context.Context, releaseService release.Interface, name string) (string, error) {
	resource, err := releaseService.Get(ctx, release.GetOptions{Name: fmt.Sprintf("v%s", name)})
	if release.IsNotFound(err) {
		return "", microerror.Maskf(releaseNotFoundError, "Release v%s could not be found.", name)
	} else if err != nil {
		return "", microerror.Mask(err)
	}

	components := resource.(*release.Release).CR.Spec.Components
	for _, component := range components {
		if component.Name == "cert-operator" {
			return component.Version, nil
		}
	}

	return "", microerror.Maskf(missingComponentError, "The release v%s does not include the required 'cert-operator' component.", name)
}

func validateProvider(provider string) error {
	if provider == key.ProviderAWS || provider == key.ProviderAzure {
		return nil
	}

	return microerror.Maskf(unsupportedProviderError, "Creating a client certificate for a workload cluster is only supported on AWS and Azure.")
}

func validateReleaseVersion(version, provider string) error {
	featureService := feature.New(provider)
	if featureService.Supports(feature.ClientCert, version) {
		return nil
	}

	if provider == key.ProviderAWS {
		return microerror.Maskf(unsupportedReleaseVersionError, "On AWS, the workload cluster must use release v13.0.0 or newer in order to allow client certificate creation.")
	}

	return microerror.Maskf(unsupportedReleaseVersionError, "The workload cluster release does not allow client certificate creation.")
}
