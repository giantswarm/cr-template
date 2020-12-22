package feature

import (
	"github.com/blang/semver"
)

// Service is used to compute different provider capabilities.
type Service struct {
	provider string
	features map[string]Feature
}

func New(provider string) *Service {
	s := &Service{
		provider: provider,
		features: map[string]Feature{
			Autoscaling:        autoscaling,
			NodePoolConditions: nodePoolConditions,
		},
	}

	return s
}

// Supports checks if a certain feature is supported or not on a given release version.
func (s *Service) Supports(featureName string, releaseVersion string) bool {
	feature, exists := s.features[featureName]
	if !exists {
		return false
	}

	capability, exists := feature[s.provider]
	if !exists {
		return false
	}

	inputVersion := semver.MustParse(releaseVersion)
	featureMinVersion := semver.MustParse(capability.MinVersion)

	return inputVersion.GE(featureMinVersion)
}
