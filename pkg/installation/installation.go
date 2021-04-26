package installation

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kubectl-gs/pkg/graphql"
)

const (
	requestTimeout = 15 * time.Second

	// management cluster internal api prefix
	internalAPIPrefix = "internal-g8s"

	urlDelimiter = "."
)

type Installation struct {
	K8sApiURL         string
	K8sInternalApiURL string
	AuthURL           string
	Provider          string
	Codename          string
	CACert            string
}

func New(ctx context.Context, fromUrl string) (*Installation, error) {
	basePath, err := getBasePath(fromUrl)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var gqlClient graphql.Client
	{
		httpClient := http.DefaultClient
		httpClient.Timeout = requestTimeout

		athenaUrl := getAthenaUrl(basePath)
		config := graphql.ClientImplConfig{
			HttpClient: httpClient,
			Url:        fmt.Sprintf("%s/graphql", athenaUrl),
		}
		gqlClient, err = graphql.NewClient(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	info, err := getInstallationInfo(ctx, gqlClient)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	baseEndpoint := strings.Split(info.Kubernetes.ApiUrl, urlDelimiter)[1:]
	k8sInternalAPI := fmt.Sprintf("https://%s.%s", internalAPIPrefix, strings.Join(baseEndpoint, urlDelimiter))
	i := &Installation{
		K8sApiURL:         info.Kubernetes.ApiUrl,
		K8sInternalApiURL: k8sInternalAPI,
		AuthURL:           info.Kubernetes.AuthUrl,
		Provider:          info.Identity.Provider,
		Codename:          info.Identity.Codename,
		CACert:            info.Kubernetes.CaCert,
	}

	return i, nil
}
