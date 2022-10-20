package clusters

import (
	"bytes"
	"testing"
	"time"

	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/google/go-cmp/cmp"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/giantswarm/kubectl-gs/internal/key"
	"github.com/giantswarm/kubectl-gs/pkg/data/domain/cluster"
	"github.com/giantswarm/kubectl-gs/pkg/output"
	"github.com/giantswarm/kubectl-gs/test/goldenfile"
)

// Test_printOutput uses golden files.
//
// go test ./cmd/get/clusters -run Test_printOutput -update
func Test_printOutputOldTemp(t *testing.T) {
	testCases := []struct {
		name               string
		clusterRes         cluster.Resource
		created            string
		provider           string
		outputType         string
		expectedGoldenFile string
	}{
		{
			name: "case 0: print list of AWS clusters, with table output",
			clusterRes: newClusterCollection(
				*newAWSClusterOT("1sad2", time.Now().Format(time.RFC3339), "12.0.0", "test", "test cluster 1", label.ServicePriorityHighest, nil),
				*newAWSClusterOT("2a03f", time.Now().Format(time.RFC3339), "11.0.0", "test", "test cluster 2", label.ServicePriorityMedium, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
				*newAWSClusterOT("asd29", time.Now().Format(time.RFC3339), "10.5.0", "test", "test cluster 3", label.ServicePriorityLowest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated, infrastructurev1alpha3.ClusterStatusConditionCreating}),
				*newAWSClusterOT("f930q", time.Now().Format(time.RFC3339), "11.0.0", "some-other", "test cluster 4", "", nil),
				*newAWSClusterOT("9f012", time.Now().Format(time.RFC3339), "9.0.0", "test", "test cluster 5", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting}),
				*newAWSClusterOT("2f0as", time.Now().Format(time.RFC3339), "10.5.0", "random", "test cluster 6", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting, infrastructurev1alpha3.ClusterStatusConditionCreated}),
			),
			provider:           key.ProviderAWS,
			outputType:         output.TypeDefault,
			expectedGoldenFile: "print_list_of_aws_clusters_table_output.golden",
		},
		{
			name: "case 1: print list of AWS clusters, with JSON output",
			clusterRes: newClusterCollection(
				*newAWSClusterOT("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test", "test cluster 1", label.ServicePriorityHighest, nil),
				*newAWSClusterOT("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test", "test cluster 2", label.ServicePriorityMedium, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
				*newAWSClusterOT("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test", "test cluster 3", label.ServicePriorityLowest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated, infrastructurev1alpha3.ClusterStatusConditionCreating}),
				*newAWSClusterOT("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test cluster 4", "", nil),
				*newAWSClusterOT("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test", "test cluster 5", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting}),
				*newAWSClusterOT("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "random", "test cluster 6", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting, infrastructurev1alpha3.ClusterStatusConditionCreated}),
			),
			provider:           key.ProviderAWS,
			outputType:         output.TypeJSON,
			expectedGoldenFile: "print_list_of_aws_clusters_json_output.golden",
		},
		{
			name: "case 2: print list of AWS clusters, with YAML output",
			clusterRes: newClusterCollection(
				*newAWSClusterOT("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test", "test cluster 1", label.ServicePriorityHighest, nil),
				*newAWSClusterOT("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test", "test cluster 2", label.ServicePriorityMedium, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
				*newAWSClusterOT("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test", "test cluster 3", label.ServicePriorityLowest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated, infrastructurev1alpha3.ClusterStatusConditionCreating}),
				*newAWSClusterOT("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test cluster 4", "", nil),
				*newAWSClusterOT("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test", "test cluster 5", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting}),
				*newAWSClusterOT("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "random", "test cluster 6", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting, infrastructurev1alpha3.ClusterStatusConditionCreated}),
			),
			created:            "2021-01-02T15:04:32Z",
			provider:           key.ProviderAWS,
			outputType:         output.TypeYAML,
			expectedGoldenFile: "print_list_of_aws_clusters_yaml_output.golden",
		},
		{
			name: "case 3: print list of AWS clusters, with name output",
			clusterRes: newClusterCollection(
				*newAWSClusterOT("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test", "test cluster 1", label.ServicePriorityHighest, nil),
				*newAWSClusterOT("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test", "test cluster 2", label.ServicePriorityMedium, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
				*newAWSClusterOT("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test", "test cluster 3", label.ServicePriorityLowest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated, infrastructurev1alpha3.ClusterStatusConditionCreating}),
				*newAWSClusterOT("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test cluster 4", "", nil),
				*newAWSClusterOT("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test", "test cluster 5", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting}),
				*newAWSClusterOT("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "random", "test cluster 6", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting, infrastructurev1alpha3.ClusterStatusConditionCreated}),
			),
			provider:           key.ProviderAWS,
			outputType:         output.TypeName,
			expectedGoldenFile: "print_list_of_aws_clusters_name_output.golden",
		},
		{
			name:               "case 4: print single AWS cluster, with table output",
			clusterRes:         newAWSClusterOT("f930q", time.Now().Format(time.RFC3339), "11.0.0", "some-other", "test cluster 4", label.ServicePriorityHighest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
			provider:           key.ProviderAWS,
			outputType:         output.TypeDefault,
			expectedGoldenFile: "print_single_aws_cluster_table_output.golden",
		},
		{
			name:               "case 5: print single AWS cluster, with JSON output",
			clusterRes:         newAWSClusterOT("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test cluster 4", label.ServicePriorityHighest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
			provider:           key.ProviderAWS,
			outputType:         output.TypeJSON,
			expectedGoldenFile: "print_single_aws_cluster_json_output.golden",
		},
		{
			name:               "case 6: print single AWS cluster, with YAML output",
			clusterRes:         newAWSClusterOT("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test cluster 4", label.ServicePriorityHighest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
			provider:           key.ProviderAWS,
			outputType:         output.TypeYAML,
			expectedGoldenFile: "print_single_aws_cluster_yaml_output.golden",
		},
		{
			name:               "case 7: print single AWS cluster, with name output",
			clusterRes:         newAWSClusterOT("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test cluster 4", label.ServicePriorityHighest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
			provider:           key.ProviderAWS,
			outputType:         output.TypeName,
			expectedGoldenFile: "print_single_aws_cluster_name_output.golden",
		},
		{
			name: "case 8: print list of Azure clusters, with table output",
			clusterRes: newClusterCollection(
				*newAzureClusterOT("1sad2", time.Now().Format(time.RFC3339), "12.0.0", "test", "test cluster 1", label.ServicePriorityHighest, nil),
				*newAzureClusterOT("2a03f", time.Now().Format(time.RFC3339), "11.0.0", "test", "test cluster 2", label.ServicePriorityMedium, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
				*newAzureClusterOT("asd29", time.Now().Format(time.RFC3339), "10.5.0", "test", "test cluster 3", label.ServicePriorityLowest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated, infrastructurev1alpha3.ClusterStatusConditionCreating}),
				*newAzureClusterOT("f930q", time.Now().Format(time.RFC3339), "11.0.0", "some-other", "test cluster 4", "", nil),
				*newAzureClusterOT("9f012", time.Now().Format(time.RFC3339), "9.0.0", "test", "test cluster 5", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting}),
				*newAzureClusterOT("2f0as", time.Now().Format(time.RFC3339), "10.5.0", "random", "test cluster 6", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting, infrastructurev1alpha3.ClusterStatusConditionCreated}),
			),
			provider:           key.ProviderAzure,
			outputType:         output.TypeDefault,
			expectedGoldenFile: "print_list_of_azure_clusters_table_output.golden",
		},
		{
			name: "case 9: print list of Azure clusters, with JSON output",
			clusterRes: newClusterCollection(
				*newAzureClusterOT("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test", "test cluster 1", label.ServicePriorityHighest, nil),
				*newAzureClusterOT("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test", "test cluster 2", label.ServicePriorityMedium, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
				*newAzureClusterOT("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test", "test cluster 3", label.ServicePriorityLowest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated, infrastructurev1alpha3.ClusterStatusConditionCreating}),
				*newAzureClusterOT("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test cluster 4", "", nil),
				*newAzureClusterOT("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test", "test cluster 5", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting}),
				*newAzureClusterOT("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "random", "test cluster 6", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting, infrastructurev1alpha3.ClusterStatusConditionCreated}),
			),
			provider:           key.ProviderAzure,
			outputType:         output.TypeJSON,
			expectedGoldenFile: "print_list_of_azure_clusters_json_output.golden",
		},
		{
			name: "case 10: print list of Azure clusters, with YAML output",
			clusterRes: newClusterCollection(
				*newAzureClusterOT("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test", "test cluster 1", label.ServicePriorityHighest, nil),
				*newAzureClusterOT("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test", "test cluster 2", label.ServicePriorityMedium, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
				*newAzureClusterOT("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test", "test cluster 3", label.ServicePriorityLowest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated, infrastructurev1alpha3.ClusterStatusConditionCreating}),
				*newAzureClusterOT("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test cluster 4", "", nil),
				*newAzureClusterOT("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test", "test cluster 5", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting}),
				*newAzureClusterOT("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "random", "test cluster 6", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting, infrastructurev1alpha3.ClusterStatusConditionCreated}),
			),
			provider:           key.ProviderAzure,
			outputType:         output.TypeYAML,
			expectedGoldenFile: "print_list_of_azure_clusters_yaml_output.golden",
		},
		{
			name: "case 11: print list of Azure clusters, with name output",
			clusterRes: newClusterCollection(
				*newAzureClusterOT("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test", "test cluster 1", label.ServicePriorityHighest, nil),
				*newAzureClusterOT("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test", "test cluster 2", label.ServicePriorityMedium, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
				*newAzureClusterOT("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test", "test cluster 3", label.ServicePriorityLowest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated, infrastructurev1alpha3.ClusterStatusConditionCreating}),
				*newAzureClusterOT("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test cluster 4", "", nil),
				*newAzureClusterOT("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test", "test cluster 5", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting}),
				*newAzureClusterOT("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "random", "test cluster 6", "", []string{infrastructurev1alpha3.ClusterStatusConditionDeleting, infrastructurev1alpha3.ClusterStatusConditionCreated}),
			),
			provider:           key.ProviderAzure,
			outputType:         output.TypeName,
			expectedGoldenFile: "print_list_of_azure_clusters_name_output.golden",
		},
		{
			name:               "case 12: print single Azure cluster, with table output",
			clusterRes:         newAzureClusterOT("f930q", time.Now().Format(time.RFC3339), "11.0.0", "some-other", "test cluster 4", label.ServicePriorityHighest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
			provider:           key.ProviderAzure,
			outputType:         output.TypeDefault,
			expectedGoldenFile: "print_single_azure_cluster_table_output.golden",
		},
		{
			name:               "case 13: print single Azure cluster, with JSON output",
			clusterRes:         newAzureClusterOT("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test cluster 4", label.ServicePriorityHighest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
			provider:           key.ProviderAzure,
			outputType:         output.TypeJSON,
			expectedGoldenFile: "print_single_azure_cluster_json_output.golden",
		},
		{
			name:               "case 14: print single Azure cluster, with YAML output",
			clusterRes:         newAzureClusterOT("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test cluster 4", label.ServicePriorityHighest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
			provider:           key.ProviderAzure,
			outputType:         output.TypeYAML,
			expectedGoldenFile: "print_single_azure_cluster_yaml_output.golden",
		},
		{
			name:               "case 15: print single Azure cluster, with name output",
			clusterRes:         newAzureClusterOT("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test cluster 4", label.ServicePriorityHighest, []string{infrastructurev1alpha3.ClusterStatusConditionCreated}),
			provider:           key.ProviderAzure,
			outputType:         output.TypeName,
			expectedGoldenFile: "print_single_azure_cluster_name_output.golden",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flag := &flag{
				print: genericclioptions.NewPrintFlags("").WithDefaultOutput(tc.outputType),
			}
			out := new(bytes.Buffer)
			runner := &runner{
				flag:     flag,
				stdout:   out,
				provider: tc.provider,
			}

			err := runner.printOutput(tc.clusterRes)

			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			var expectedResult []byte
			{
				gf := goldenfile.New("testdata", tc.expectedGoldenFile)
				if *update {
					err = gf.Update(out.Bytes())
					if err != nil {
						t.Fatalf("unexpected error: %s", err.Error())
					}
					expectedResult = out.Bytes()
				} else {
					expectedResult, err = gf.Read()
					if err != nil {
						t.Fatalf("unexpected error: %s", err.Error())
					}
				}
			}

			diff := cmp.Diff(string(expectedResult), out.String())
			if diff != "" {
				t.Fatalf("value not expected, got:\n %s", diff)
			}
		})
	}
}

func newcapiClusterOT(id, created, release, org, description, servicePriority string, conditions []string) *capi.Cluster {
	return newcapiCluster(id, release, org, description, servicePriority, parseCreated(created), conditions)
}

func newAWSClusterResourceOT(id, created, release, org, description string, conditions []string) *infrastructurev1alpha3.AWSCluster {
	return newAWSClusterResource(id, release, org, description, parseCreated(created), conditions)
}

func newAWSClusterOT(id, created, release, org, description, servicePriority string, conditions []string) *cluster.Cluster {
	return newAWSCluster(id, release, org, description, servicePriority, parseCreated(created), conditions)
}

func newAzureClusterOT(id, created, release, org, description, servicePriority string, conditions []string) *cluster.Cluster {
	return newAzureCluster(id, release, org, description, servicePriority, parseCreated(created), conditions)
}

func Test_printNoResourcesOutputOldTemp(t *testing.T) {
	expected := `No clusters found.
To create a cluster, please check

  kubectl gs template cluster --help
`
	out := new(bytes.Buffer)
	runner := &runner{
		stdout: out,
	}
	runner.printNoResourcesOutput()

	if out.String() != expected {
		t.Fatalf("value not expected, got:\n %s", out.String())
	}
}
