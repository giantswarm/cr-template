package app

import (
	"context"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/application/v1alpha1"
	"github.com/qri-io/jsonschema"
)

const (
	defaultNamespace = "default"
)

type ValidateOptions struct {
	AllNamespaces bool
	LabelSelector string
	Name          string
	Namespace     string
	ValuesSchema  *jsonschema.Schema
}

// ValidationResults contains multiple validation results. The printer takes
// this and makes a table out of it.
type ValidationResults []*ValidationResult

// ValidationResult contains everything we need to show information about a
// validation attempt.
type ValidationResult struct {
	App applicationv1alpha1.App

	// The schema.values.json file, unmarshalled, from the
	// 'application.giantswarm.io/values-schema' annotation.
	ValuesSchema *jsonschema.Schema

	// An array of validation errors that surfaced after validating the merged
	// values against the schema.values.json file. In the context of this struct,
	// this is the money maker. This is what we really care about, it holds the
	// actual validation errors (if any) that we want to show the user.
	ValidationErrors []jsonschema.KeyError

	// Any error that occured while attempting to validate the values of this
	// app. This is not a validation error, this is any actual error occured
	// while trying to gather all the files and information required to make a
	// validation pass. An error here means we were not able to validate the app.
	Err error
}

type CatalogFetchResult struct {
	catalog *applicationv1alpha1.AppCatalog
	index   *IndexFile

	err error
}

type SchemaFetchResult struct {
	schema *jsonschema.Schema
	err    error
}

type IndexFile struct {
	APIVersion string                   `yaml:"apiVersion"`
	Entries    map[string]ChartVersions `yaml:"entries"`
}

// ChartVersions is a list of versioned chart references.
type ChartVersions []*ChartVersion

// ChartVersion represents a chart entry in the IndexFile
type ChartVersion struct {
	Name        string            `json:"name,omitempty"`
	Version     string            `json:"version,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	URLs        []string          `json:"urls"`
}

// Interface represents the contract for the apps service.
// Using this instead of a regular 'struct' makes mocking the
// service in tests much simpler.
type Interface interface {
	// GetAllInNamespace(context.Context, GetAllInNamespaceOptions) (runtime.Object, error)
	// GetByName(context.Context, GetByNameOptions) (runtime.Object, error)
	// GetAll(context.Contect, GetAllOptions) (runtime.Object, error)
	Validate(context.Context, ValidateOptions) (ValidationResults, error)
}
