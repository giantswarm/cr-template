package nodepools

import (
	"bytes"
	goflag "flag"
	"testing"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/kubectl-gs/internal/key"
	"github.com/giantswarm/kubectl-gs/internal/label"
	"github.com/giantswarm/kubectl-gs/pkg/data/domain/nodepool"
	"github.com/giantswarm/kubectl-gs/pkg/output"
	"github.com/giantswarm/kubectl-gs/test/goldenfile"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	capiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

var update = goflag.Bool("update", false, "update .golden reference test files")

// Test_printOutput uses golden files.
//
//  go test ./cmd/get/nodepools -run Test_printOutput -update
//
func Test_printOutput(t *testing.T) {
	testCases := []struct {
		name               string
		np                 nodepool.Resource
		provider           string
		outputType         string
		expectedGoldenFile string
	}{
		{
			name: "case 0: print list of AWS nodepools, with table output",
			np: newNodePoolCollection(
				*newAWSNodePool("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test nodepool 1", 1, 3, 2, 2),
				*newAWSNodePool("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 2", 3, 10, 5, 2),
				*newAWSNodePool("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 3", 10, 10, 10, 10),
				*newAWSNodePool("f930q", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
				*newAWSNodePool("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test nodepool 5", 0, 3, 1, 1),
				*newAWSNodePool("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 6", 2, 5, 5, 5),
			),
			provider:           key.ProviderAWS,
			outputType:         output.TypeDefault,
			expectedGoldenFile: "print_list_of_aws_nodepools_table_output.golden",
		},
		{
			name: "case 1: print list of AWS nodepools, with JSON output",
			np: newNodePoolCollection(
				*newAWSNodePool("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test nodepool 1", 1, 3, 2, 2),
				*newAWSNodePool("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 2", 3, 10, 5, 2),
				*newAWSNodePool("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 3", 10, 10, 10, 10),
				*newAWSNodePool("f930q", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
				*newAWSNodePool("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test nodepool 5", 0, 3, 1, 1),
				*newAWSNodePool("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 6", 2, 5, 5, 5),
			),
			provider:           key.ProviderAWS,
			outputType:         output.TypeJSON,
			expectedGoldenFile: "print_list_of_aws_nodepools_json_output.golden",
		},
		{
			name: "case 2: print list of AWS nodepools, with YAML output",
			np: newNodePoolCollection(
				*newAWSNodePool("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test nodepool 1", 1, 3, 2, 2),
				*newAWSNodePool("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 2", 3, 10, 5, 2),
				*newAWSNodePool("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 3", 10, 10, 10, 10),
				*newAWSNodePool("f930q", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
				*newAWSNodePool("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test nodepool 5", 0, 3, 1, 1),
				*newAWSNodePool("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 6", 2, 5, 5, 5),
			),
			provider:           key.ProviderAWS,
			outputType:         output.TypeYAML,
			expectedGoldenFile: "print_list_of_aws_nodepools_yaml_output.golden",
		},
		{
			name: "case 3: print list of AWS nodepools, with name output",
			np: newNodePoolCollection(
				*newAWSNodePool("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test nodepool 1", 1, 3, 2, 2),
				*newAWSNodePool("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 2", 3, 10, 5, 2),
				*newAWSNodePool("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 3", 10, 10, 10, 10),
				*newAWSNodePool("f930q", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
				*newAWSNodePool("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test nodepool 5", 0, 3, 1, 1),
				*newAWSNodePool("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 6", 2, 5, 5, 5),
			),
			provider:           key.ProviderAWS,
			outputType:         output.TypeName,
			expectedGoldenFile: "print_list_of_aws_nodepools_name_output.golden",
		},
		{
			name:               "case 4: print single AWS nodepool, with table output",
			np:                 newAWSNodePool("f930q", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
			provider:           key.ProviderAWS,
			outputType:         output.TypeDefault,
			expectedGoldenFile: "print_single_aws_nodepool_table_output.golden",
		},
		{
			name:               "case 5: print single AWS nodepool, with JSON output",
			np:                 newAWSNodePool("f930q", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
			provider:           key.ProviderAWS,
			outputType:         output.TypeJSON,
			expectedGoldenFile: "print_single_aws_nodepool_json_output.golden",
		},
		{
			name:               "case 6: print single AWS nodepool, with YAML output",
			np:                 newAWSNodePool("f930q", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
			provider:           key.ProviderAWS,
			outputType:         output.TypeYAML,
			expectedGoldenFile: "print_single_aws_nodepool_yaml_output.golden",
		},
		{
			name:               "case 7: print single AWS nodepool, with name output",
			np:                 newAWSNodePool("f930q", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
			provider:           key.ProviderAWS,
			outputType:         output.TypeName,
			expectedGoldenFile: "print_single_aws_nodepool_name_output.golden",
		},
		// {
		// 	name: "case 8: print list of Azure clusters, with table output",
		// 	np: newNodePoolCollection(
		// 		*newAzureCluster("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test", "test nodepool 1", nil),
		// 		*newAzureCluster("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test", "test nodepool 2", []string{infrastructurev1alpha2.ClusterStatusConditionCreated}),
		// 		*newAzureCluster("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test", "test nodepool 3", []string{infrastructurev1alpha2.ClusterStatusConditionCreated, infrastructurev1alpha2.ClusterStatusConditionCreating}),
		// 		*newAzureCluster("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test nodepool 4", nil),
		// 		*newAzureCluster("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test", "test nodepool 5", []string{infrastructurev1alpha2.ClusterStatusConditionDeleting}),
		// 		*newAzureCluster("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "random", "test nodepool 6", []string{infrastructurev1alpha2.ClusterStatusConditionDeleting, infrastructurev1alpha2.ClusterStatusConditionCreated}),
		// 	),
		// 	provider:           key.ProviderAzure,
		// 	outputType:         output.TypeDefault,
		// 	expectedGoldenFile: "print_list_of_azure_clusters_table_output.golden",
		// },
		// {
		// 	name: "case 9: print list of Azure clusters, with JSON output",
		// 	np: newNodePoolCollection(
		// 		*newAzureCluster("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test", "test nodepool 1", nil),
		// 		*newAzureCluster("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test", "test nodepool 2", []string{infrastructurev1alpha2.ClusterStatusConditionCreated}),
		// 		*newAzureCluster("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test", "test nodepool 3", []string{infrastructurev1alpha2.ClusterStatusConditionCreated, infrastructurev1alpha2.ClusterStatusConditionCreating}),
		// 		*newAzureCluster("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test nodepool 4", nil),
		// 		*newAzureCluster("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test", "test nodepool 5", []string{infrastructurev1alpha2.ClusterStatusConditionDeleting}),
		// 		*newAzureCluster("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "random", "test nodepool 6", []string{infrastructurev1alpha2.ClusterStatusConditionDeleting, infrastructurev1alpha2.ClusterStatusConditionCreated}),
		// 	),
		// 	provider:           key.ProviderAzure,
		// 	outputType:         output.TypeJSON,
		// 	expectedGoldenFile: "print_list_of_azure_clusters_json_output.golden",
		// },
		// {
		// 	name: "case 10: print list of Azure clusters, with YAML output",
		// 	np: newNodePoolCollection(
		// 		*newAzureCluster("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test", "test nodepool 1", nil),
		// 		*newAzureCluster("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test", "test nodepool 2", []string{infrastructurev1alpha2.ClusterStatusConditionCreated}),
		// 		*newAzureCluster("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test", "test nodepool 3", []string{infrastructurev1alpha2.ClusterStatusConditionCreated, infrastructurev1alpha2.ClusterStatusConditionCreating}),
		// 		*newAzureCluster("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test nodepool 4", nil),
		// 		*newAzureCluster("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test", "test nodepool 5", []string{infrastructurev1alpha2.ClusterStatusConditionDeleting}),
		// 		*newAzureCluster("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "random", "test nodepool 6", []string{infrastructurev1alpha2.ClusterStatusConditionDeleting, infrastructurev1alpha2.ClusterStatusConditionCreated}),
		// 	),
		// 	provider:           key.ProviderAzure,
		// 	outputType:         output.TypeYAML,
		// 	expectedGoldenFile: "print_list_of_azure_clusters_yaml_output.golden",
		// },
		// {
		// 	name: "case 11: print list of Azure clusters, with name output",
		// 	np: newNodePoolCollection(
		// 		*newAzureCluster("1sad2", "2021-01-02T15:04:32Z", "12.0.0", "test", "test nodepool 1", nil),
		// 		*newAzureCluster("2a03f", "2021-01-02T15:04:32Z", "11.0.0", "test", "test nodepool 2", []string{infrastructurev1alpha2.ClusterStatusConditionCreated}),
		// 		*newAzureCluster("asd29", "2021-01-02T15:04:32Z", "10.5.0", "test", "test nodepool 3", []string{infrastructurev1alpha2.ClusterStatusConditionCreated, infrastructurev1alpha2.ClusterStatusConditionCreating}),
		// 		*newAzureCluster("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test nodepool 4", nil),
		// 		*newAzureCluster("9f012", "2021-01-02T15:04:32Z", "9.0.0", "test", "test nodepool 5", []string{infrastructurev1alpha2.ClusterStatusConditionDeleting}),
		// 		*newAzureCluster("2f0as", "2021-01-02T15:04:32Z", "10.5.0", "random", "test nodepool 6", []string{infrastructurev1alpha2.ClusterStatusConditionDeleting, infrastructurev1alpha2.ClusterStatusConditionCreated}),
		// 	),
		// 	provider:           key.ProviderAzure,
		// 	outputType:         output.TypeName,
		// 	expectedGoldenFile: "print_list_of_azure_clusters_name_output.golden",
		// },
		// {
		// 	name:               "case 12: print single Azure cluster, with table output",
		// 	np:                 newAzureCluster("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test nodepool 4", []string{infrastructurev1alpha2.ClusterStatusConditionCreated}),
		// 	provider:           key.ProviderAzure,
		// 	outputType:         output.TypeDefault,
		// 	expectedGoldenFile: "print_single_azure_cluster_table_output.golden",
		// },
		// {
		// 	name:               "case 13: print single Azure cluster, with JSON output",
		// 	np:                 newAzureCluster("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test nodepool 4", []string{infrastructurev1alpha2.ClusterStatusConditionCreated}),
		// 	provider:           key.ProviderAzure,
		// 	outputType:         output.TypeJSON,
		// 	expectedGoldenFile: "print_single_azure_cluster_json_output.golden",
		// },
		// {
		// 	name:               "case 14: print single Azure cluster, with YAML output",
		// 	np:                 newAzureCluster("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test nodepool 4", []string{infrastructurev1alpha2.ClusterStatusConditionCreated}),
		// 	provider:           key.ProviderAzure,
		// 	outputType:         output.TypeYAML,
		// 	expectedGoldenFile: "print_single_azure_cluster_yaml_output.golden",
		// },
		// {
		// 	name:               "case 15: print single Azure cluster, with name output",
		// 	np:                 newAzureCluster("f930q", "2021-01-02T15:04:32Z", "11.0.0", "some-other", "test nodepool 4", []string{infrastructurev1alpha2.ClusterStatusConditionCreated}),
		// 	provider:           key.ProviderAzure,
		// 	outputType:         output.TypeName,
		// 	expectedGoldenFile: "print_single_azure_cluster_name_output.golden",
		// },
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

			err := runner.printOutput(tc.np)
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

func newAWSMachineDeployment(id, created, release, description string, nodesMin, nodesMax int) *infrastructurev1alpha2.AWSMachineDeployment {
	location, _ := time.LoadLocation("UTC")
	parsedCreationDate, _ := time.ParseInLocation(time.RFC3339, created, location)
	c := &infrastructurev1alpha2.AWSMachineDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              id,
			Namespace:         "default",
			CreationTimestamp: metav1.NewTime(parsedCreationDate),
			Labels: map[string]string{
				label.ReleaseVersion: release,
				label.Organization:   "giantswarm",
			},
		},
		Spec: infrastructurev1alpha2.AWSMachineDeploymentSpec{
			NodePool: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePool{
				Description: description,
				Scaling: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolScaling{
					Min: nodesMin,
					Max: nodesMax,
				},
			},
		},
	}

	c.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   infrastructurev1alpha2.SchemeGroupVersion.Group,
		Version: infrastructurev1alpha2.SchemeGroupVersion.Version,
		Kind:    infrastructurev1alpha2.NewAWSMachineDeploymentTypeMeta().Kind,
	})

	return c
}

func newCAPIv1alpha2MachineDeployment(id, created, release string, nodesDesired, nodesReady int) *capiv1alpha2.MachineDeployment {
	location, _ := time.LoadLocation("UTC")
	parsedCreationDate, _ := time.ParseInLocation(time.RFC3339, created, location)
	c := &capiv1alpha2.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1alpha2",
			Kind:       "MachineDeployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              id,
			Namespace:         "default",
			CreationTimestamp: metav1.NewTime(parsedCreationDate),
			Labels: map[string]string{
				label.ReleaseVersion: release,
				label.Organization:   "giantswarm",
			},
		},
		Status: capiv1alpha2.MachineDeploymentStatus{
			Replicas:      int32(nodesDesired),
			ReadyReplicas: int32(nodesReady),
		},
	}

	return c
}

func newAWSNodePool(id, created, release, description string, nodesMin, nodesMax, nodesDesired, nodesReady int) *nodepool.Nodepool {
	awsMD := newAWSMachineDeployment(id, created, release, description, nodesMin, nodesMax)
	md := newCAPIv1alpha2MachineDeployment(id, created, release, nodesDesired, nodesReady)

	np := &nodepool.Nodepool{
		MachineDeployment:    md,
		AWSMachineDeployment: awsMD,
	}

	return np
}

func newNodePoolCollection(nps ...nodepool.Nodepool) *nodepool.Collection {
	collection := &nodepool.Collection{
		Items: nps,
	}

	return collection
}
