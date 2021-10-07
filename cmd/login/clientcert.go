package login

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/kubectl-gs/pkg/data/domain/clientcert"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	credentialRetryTimeout    = 500 * time.Millisecond
	credentialMaxRetryTimeout = 20 * time.Second

	credentialKeyCertCRT = "crt"
	credentialKeyCertKey = "key"
	credentialKeyCertCA  = "ca"
)

func generateClientCertUID() string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(time.Now().String()))

	uid := fmt.Sprintf("%x", hash.Sum(nil))

	return uid[:16]
}

func generateClientCert(cluster, organization, ttl string, groups []string, clusterBasePath string) (*clientcert.ClientCert, error) {
	clientCertUID := generateClientCertUID()
	clientCertName := fmt.Sprintf("%s-%s", cluster, clientCertUID)
	commonName := fmt.Sprintf("%s.%s.k8s.%s", clientCertUID, cluster, clusterBasePath)

	certConfig := &corev1alpha1.CertConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clientCertName,
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				"cert-operator.giantswarm.io/version": "0.1.0",
				"giantswarm.io/certificate":           clientCertUID,
				"giantswarm.io/cluster":               cluster,
				"giantswarm.io/organization":          organization,
			},
		},
		Spec: corev1alpha1.CertConfigSpec{
			Cert: corev1alpha1.CertConfigSpecCert{
				AllowBareDomains:    true,
				ClusterComponent:    clientCertUID,
				ClusterID:           cluster,
				CommonName:          commonName,
				DisableRegeneration: true,
				Organizations:       groups,
				TTL:                 ttl,
			},
			VersionBundle: corev1alpha1.CertConfigSpecVersionBundle{
				Version: "0.1.0",
			},
		},
	}

	r := &clientcert.ClientCert{
		CertConfig: certConfig,
	}

	return r, nil
}

// tryToGetClientCertCredential tries to fetch the client certificate credential
// for a couple of times, until the resource is fetched, or until the timeout is reached.
func tryToGetClientCertCredential(ctx context.Context, clientCertService clientcert.Interface, name string) (*corev1.Secret, error) {
	var secret *corev1.Secret
	var err error

	o := func() error {
		secret, err = clientCertService.GetCredential(ctx, name)
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

// storeClientCertCredential saves the created client certificate credentials into the kubectl config.
func storeClientCertCredential(k8sConfigAccess clientcmd.ConfigAccess, fs afero.Fs, clientCert *clientcert.ClientCert, credential *corev1.Secret, clusterBasePath string) (string, error) {
	config, err := k8sConfigAccess.GetStartingConfig()
	if err != nil {
		return "", microerror.Mask(err)
	}

	contextName := clientCert.CertConfig.GetName()
	userName := clientCert.CertConfig.Spec.Cert.ClusterComponent
	clusterName := fmt.Sprintf("giantswarm-%s", clientCert.CertConfig.Spec.Cert.ClusterID)
	clusterServer := fmt.Sprintf("https://api.%s.k8s.%s", clientCert.CertConfig.Spec.Cert.ClusterID, clusterBasePath)

	certCRT := credential.Data[credentialKeyCertCRT]
	certKey := credential.Data[credentialKeyCertKey]
	certCA := credential.Data[credentialKeyCertCA]

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
		context, exists := config.Contexts[contextName]
		if !exists {
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
		return "", microerror.Mask(err)
	}

	return contextName, nil
}

func cleanUpClientCertResources(ctx context.Context, clientCertService clientcert.Interface, clientCertResource *clientcert.ClientCert) error {
	err := clientCertService.Delete(ctx, clientCertResource)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
