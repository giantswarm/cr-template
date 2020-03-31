package key

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
)

const (
	// IDChars represents the character set used to generate cluster IDs.
	// (does not contain 1 and l, to avoid confusion)
	idChars = "023456789abcdefghijkmnopqrstuvwxyz"
	// IDLength represents the number of characters used to create a cluster ID.
	idLength = 5
)

func GenerateID() string {
	for {
		letterRunes := []rune(idChars)
		b := make([]rune, idLength)
		rand.Seed(time.Now().UnixNano())
		for i := range b {
			b[i] = letterRunes[rand.Intn(len(letterRunes))]
		}

		id := string(b)

		if _, err := strconv.Atoi(id); err == nil {
			// string is numbers only, which we want to avoid
			continue
		}

		matched, err := regexp.MatchString("^[a-z]+$", id)
		if err == nil && matched == true {
			// strings is letters only, which we also avoid
			continue
		}

		return id
	}
}

func GenerateAssetName(values ...string) string {
	return strings.Join(values, "-")
}

// readConfigMapFromFile reads a configmap from a YAML file.
func ReadConfigMapYamlFromFile(fs afero.Fs, path string) (string, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return "", microerror.Mask(err)
	}

	rawMap := map[string]string{}
	err = yaml.Unmarshal(data, &rawMap)
	if err != nil {
		return "", microerror.Maskf(unmashalToMapFailedError, err.Error())
	}

	return string(data), nil
}

// readSecretFromFile reads a configmap from a YAML file.
func ReadSecretYamlFromFile(fs afero.Fs, path string) (map[string][]byte, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	rawMap := map[string][]byte{}
	err = yaml.Unmarshal(data, &rawMap)
	if err != nil {
		return nil, microerror.Maskf(unmashalToMapFailedError, err.Error())
	}

	return rawMap, nil
}
