package installation

import (
	"testing"

	"github.com/pkg/errors"
)

func Test_GetBasePath(t *testing.T) {
	testCases := []struct {
		name                   string
		url                    string
		expectedResult         string
		expectedInternalResult string
		errorMatcher           func(error) bool
	}{
		{
			name:                   "case 0: using k8s api url",
			url:                    "https://g8s.test.eu-west-1.aws.coolio.com",
			expectedResult:         "g8s.test.eu-west-1.aws.coolio.com",
			expectedInternalResult: "internal-g8s.test.eu-west-1.aws.coolio.com",
		},
		{
			name:                   "case 1: using k8s api url, without scheme",
			url:                    "g8s.test.eu-west-1.aws.coolio.com",
			expectedResult:         "g8s.test.eu-west-1.aws.coolio.com",
			expectedInternalResult: "internal-g8s.test.eu-west-1.aws.coolio.com",
		},
		{
			name:                   "case 2: using k8s api url, with trailing slash",
			url:                    "g8s.test.eu-west-1.aws.coolio.com/",
			expectedResult:         "g8s.test.eu-west-1.aws.coolio.com",
			expectedInternalResult: "internal-g8s.test.eu-west-1.aws.coolio.com",
		},
		{
			name:                   "case 3: using happa url",
			url:                    "happa.g8s.test.eu-west-1.aws.coolio.com/",
			expectedResult:         "g8s.test.eu-west-1.aws.coolio.com",
			expectedInternalResult: "internal-g8s.test.eu-west-1.aws.coolio.com",
		},
		{
			name:         "case 4: using invalid url",
			url:          "coolio.com",
			errorMatcher: IsUnknownUrlType,
		},
		{
			name:         "case 4: using empty url",
			url:          "",
			errorMatcher: IsUnknownUrlType,
		},
		{
			name:                   "case 5: using giant swarm api url",
			url:                    "https://api.g8s.test.eu-west-1.aws.coolio.com",
			expectedResult:         "g8s.test.eu-west-1.aws.coolio.com",
			expectedInternalResult: "internal-g8s.test.eu-west-1.aws.coolio.com",
		},
		{
			name:                   "case 5: new style url",
			url:                    "https://api.installation.customer.gigantic.io:6443",
			expectedResult:         "installation.customer.gigantic.io",
			expectedInternalResult: "internal-installation.customer.gigantic.io",
		},
		{
			name:                   "case 6: wc k8s api url",
			url:                    "https://api.abc12.k8s.installation.customer.gigantic.io:6443",
			expectedResult:         "abc12.k8s.installation.customer.gigantic.io",
			expectedInternalResult: "internal-api.abc12.k8s.installation.customer.gigantic.io",
		},
		{
			name:                   "case 6: wc k8s api long url",
			url:                    "https://api.abc12def34.k8s.installation.customer.gigantic.io:6443",
			expectedResult:         "abc12def34.k8s.installation.customer.gigantic.io",
			expectedInternalResult: "internal-api.abc12def34.k8s.installation.customer.gigantic.io",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			basePath, internalPath, err := GetBaseAndInternalPath(tc.url)
			if tc.errorMatcher != nil {
				if !tc.errorMatcher(err) {
					t.Fatalf("error not matching expected matcher, got: %s", errors.Cause(err))
				}

				// All good. Fall through.
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			if basePath != tc.expectedResult {
				t.Fatalf("base path not expected, got: %s, want: %s", basePath, tc.expectedResult)
			}
			if internalPath != tc.expectedInternalResult {
				t.Fatalf("internal path not expected, got: %s, want: %s", internalPath, tc.expectedInternalResult)
			}
		})
	}
}

func Test_getAthenaUrl(t *testing.T) {
	basePath := "g8s.test.eu-west-1.aws.coolio.com" // nolint:goconst
	expectedResult := "https://athena.g8s.test.eu-west-1.aws.coolio.com"

	if result := getAthenaUrl(basePath); result != expectedResult {
		t.Fatalf("url not expected, got: %s", result)
	}
}
