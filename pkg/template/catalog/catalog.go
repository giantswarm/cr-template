package catalog

import (
	applicationv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Config struct {
	CatalogConfigMapName string
	CatalogSecretName    string
	Description          string
	ID                   string
	Name                 string
	Namespace            string
	LogoURL              string
	URL                  string
}

func NewCatalogCR(config Config) (*applicationv1alpha1.Catalog, error) {
	var catalogConfig *applicationv1alpha1.CatalogSpecConfig

	if config.CatalogConfigMapName != "" || config.CatalogSecretName != "" {
		catalogConfig = &applicationv1alpha1.CatalogSpecConfig{}
	}

	if config.CatalogConfigMapName != "" {
		catalogConfig.ConfigMap = &applicationv1alpha1.CatalogSpecConfigConfigMap{
			Name:      config.CatalogConfigMapName,
			Namespace: config.Namespace,
		}
	}

	if config.CatalogSecretName != "" {
		catalogConfig.Secret = &applicationv1alpha1.CatalogSpecConfigSecret{
			Name:      config.CatalogSecretName,
			Namespace: config.Namespace,
		}
	}

	catalogCR := &applicationv1alpha1.Catalog{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Catalog",
			APIVersion: "application.giantswarm.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
			Labels: map[string]string{
				"app-operator.giantswarm.io/version":     "1.0.0",
				"application.giantswarm.io/catalog-type": "awesome",
			},
		},
		Spec: applicationv1alpha1.CatalogSpec{
			Config:      catalogConfig,
			Description: config.Description,
			LogoURL:     config.LogoURL,
			Storage: applicationv1alpha1.CatalogSpecStorage{
				URL:  config.URL,
				Type: "helm",
			},
			Title: config.Name,
		},
	}

	return catalogCR, nil
}

func NewConfigMap(config Config, data string) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.CatalogConfigMapName,
			Namespace: config.Namespace,
		},
		Data: map[string]string{
			"values": data,
		},
	}

	return configMap, nil
}

func NewSecret(config Config, data []byte) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.CatalogSecretName,
			Namespace: config.Namespace,
		},
		Data: map[string][]byte{
			"values": data,
		},
	}

	return secret, nil
}
