package nodepools

import (
	"bytes"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/giantswarm/kubectl-gs/internal/key"
	"github.com/giantswarm/kubectl-gs/pkg/data/domain/nodepool"
	"github.com/giantswarm/kubectl-gs/pkg/output"
	"github.com/giantswarm/kubectl-gs/test/goldenfile"
)

// Test_printOutput uses golden files.
//
//	go test ./cmd/get/nodepools -run Test_printOutput -update
func Test_printOutputOldTemp(t *testing.T) {
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
				*newAWSNodePoolOT("1sad2", "s921a", time.Now().Format(time.RFC3339), "12.0.0", "test nodepool 1", 1, 3, 2, 2),
				*newAWSNodePoolOT("2a03f", "3a0d1", time.Now().Format(time.RFC3339), "11.0.0", "test nodepool 2", 3, 10, 5, 2),
				*newAWSNodePoolOT("asd29", "s0a10", time.Now().Format(time.RFC3339), "10.5.0", "test nodepool 3", 10, 10, 10, 10),
				*newAWSNodePoolOT("f930q", "s921a", time.Now().Format(time.RFC3339), "11.0.0", "test nodepool 4", 3, 3, 3, 1),
				*newAWSNodePoolOT("9f012", "29sa0", time.Now().Format(time.RFC3339), "9.0.0", "test nodepool 5", 0, 3, 1, 1),
				*newAWSNodePoolOT("2f0as", "s00sn", time.Now().Format(time.RFC3339), "10.5.0", "test nodepool 6", 2, 5, 5, 5),
			),
			provider:           key.ProviderAWS,
			outputType:         output.TypeDefault,
			expectedGoldenFile: "print_list_of_aws_nodepools_table_output.golden",
		},
		{
			name: "case 1: print list of AWS nodepools, with JSON output",
			np: newNodePoolCollection(
				*newAWSNodePoolOT("1sad2", "s921a", "2021-01-02T15:04:32Z", "12.0.0", "test nodepool 1", 1, 3, 2, 2),
				*newAWSNodePoolOT("2a03f", "3a0d1", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 2", 3, 10, 5, 2),
				*newAWSNodePoolOT("asd29", "s0a10", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 3", 10, 10, 10, 10),
				*newAWSNodePoolOT("f930q", "s921a", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
				*newAWSNodePoolOT("9f012", "29sa0", "2021-01-02T15:04:32Z", "9.0.0", "test nodepool 5", 0, 3, 1, 1),
				*newAWSNodePoolOT("2f0as", "s00sn", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 6", 2, 5, 5, 5),
			),
			provider:           key.ProviderAWS,
			outputType:         output.TypeJSON,
			expectedGoldenFile: "print_list_of_aws_nodepools_json_output.golden",
		},
		{
			name: "case 2: print list of AWS nodepools, with YAML output",
			np: newNodePoolCollection(
				*newAWSNodePoolOT("1sad2", "s921a", "2021-01-02T15:04:32Z", "12.0.0", "test nodepool 1", 1, 3, 2, 2),
				*newAWSNodePoolOT("2a03f", "3a0d1", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 2", 3, 10, 5, 2),
				*newAWSNodePoolOT("asd29", "s0a10", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 3", 10, 10, 10, 10),
				*newAWSNodePoolOT("f930q", "s921a", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
				*newAWSNodePoolOT("9f012", "29sa0", "2021-01-02T15:04:32Z", "9.0.0", "test nodepool 5", 0, 3, 1, 1),
				*newAWSNodePoolOT("2f0as", "s00sn", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 6", 2, 5, 5, 5),
			),
			provider:           key.ProviderAWS,
			outputType:         output.TypeYAML,
			expectedGoldenFile: "print_list_of_aws_nodepools_yaml_output.golden",
		},
		{
			name: "case 3: print list of AWS nodepools, with name output",
			np: newNodePoolCollection(
				*newAWSNodePoolOT("1sad2", "s921a", "2021-01-02T15:04:32Z", "12.0.0", "test nodepool 1", 1, 3, 2, 2),
				*newAWSNodePoolOT("2a03f", "3a0d1", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 2", 3, 10, 5, 2),
				*newAWSNodePoolOT("asd29", "s0a10", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 3", 10, 10, 10, 10),
				*newAWSNodePoolOT("f930q", "s921a", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
				*newAWSNodePoolOT("9f012", "29sa0", "2021-01-02T15:04:32Z", "9.0.0", "test nodepool 5", 0, 3, 1, 1),
				*newAWSNodePoolOT("2f0as", "s00sn", "2021-01-02T15:04:32Z", "10.5.0", "test nodepool 6", 2, 5, 5, 5),
			),
			provider:           key.ProviderAWS,
			outputType:         output.TypeName,
			expectedGoldenFile: "print_list_of_aws_nodepools_name_output.golden",
		},
		{
			name:               "case 4: print single AWS nodepool, with table output",
			np:                 newAWSNodePoolOT("f930q", "s921a", time.Now().Format(time.RFC3339), "11.0.0", "test nodepool 4", 3, 3, 3, 1),
			provider:           key.ProviderAWS,
			outputType:         output.TypeDefault,
			expectedGoldenFile: "print_single_aws_nodepool_table_output.golden",
		},
		{
			name:               "case 5: print single AWS nodepool, with JSON output",
			np:                 newAWSNodePoolOT("f930q", "s921a", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
			provider:           key.ProviderAWS,
			outputType:         output.TypeJSON,
			expectedGoldenFile: "print_single_aws_nodepool_json_output.golden",
		},
		{
			name:               "case 6: print single AWS nodepool, with YAML output",
			np:                 newAWSNodePoolOT("f930q", "s921a", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
			provider:           key.ProviderAWS,
			outputType:         output.TypeYAML,
			expectedGoldenFile: "print_single_aws_nodepool_yaml_output.golden",
		},
		{
			name:               "case 7: print single AWS nodepool, with name output",
			np:                 newAWSNodePoolOT("f930q", "s921a", "2021-01-02T15:04:32Z", "11.0.0", "test nodepool 4", 3, 3, 3, 1),
			provider:           key.ProviderAWS,
			outputType:         output.TypeName,
			expectedGoldenFile: "print_single_aws_nodepool_name_output.golden",
		},
		{
			name: "case 8: print list of Azure nodepools, with table output",
			np: newNodePoolCollection(
				*newAzureNodePoolOT("1sad2", "s921a", time.Now().Format(time.RFC3339), "13.0.0", "test nodepool 1", 1, 3, -1, -1),
				*newAzureNodePoolOT("2a03f", "3a0d1", time.Now().Format(time.RFC3339), "13.0.0", "test nodepool 2", 3, 10, -1, -1),
				*newAzureNodePoolOT("asd29", "s0a10", time.Now().Format(time.RFC3339), "13.2.0", "test nodepool 3", 10, 10, 10, 10),
				*newAzureNodePoolOT("f930q", "s921a", time.Now().Format(time.RFC3339), "13.0.0", "test nodepool 4", 3, 3, -1, -1),
				*newAzureNodePoolOT("9f012", "29sa0", time.Now().Format(time.RFC3339), "13.2.0", "test nodepool 5", 0, 3, 1, 1),
				*newAzureNodePoolOT("2f0as", "s00sn", time.Now().Format(time.RFC3339), "13.1.0", "test nodepool 6", 2, 5, -1, -1),
			),
			provider:           key.ProviderAzure,
			outputType:         output.TypeDefault,
			expectedGoldenFile: "print_list_of_azure_nodepools_table_output.golden",
		},
		{
			name: "case 9: print list of Azure nodepools, with JSON output",
			np: newNodePoolCollection(
				*newAzureNodePoolOT("1sad2", "s921a", "2021-01-02T15:04:32Z", "13.0.0", "test nodepool 1", 1, 3, -1, -1),
				*newAzureNodePoolOT("2a03f", "3a0d1", "2021-01-02T15:04:32Z", "13.0.0", "test nodepool 2", 3, 10, -1, -1),
				*newAzureNodePoolOT("asd29", "s0a10", "2021-01-02T15:04:32Z", "13.2.0", "test nodepool 3", 10, 10, 10, 10),
				*newAzureNodePoolOT("f930q", "s921a", "2021-01-02T15:04:32Z", "13.0.0", "test nodepool 4", 3, 3, -1, -1),
				*newAzureNodePoolOT("9f012", "29sa0", "2021-01-02T15:04:32Z", "13.2.0", "test nodepool 5", 0, 3, 1, 1),
				*newAzureNodePoolOT("2f0as", "s00sn", "2021-01-02T15:04:32Z", "13.1.0", "test nodepool 6", 2, 5, -1, -1),
			),
			provider:           key.ProviderAzure,
			outputType:         output.TypeJSON,
			expectedGoldenFile: "print_list_of_azure_nodepools_json_output.golden",
		},
		{
			name: "case 10: print list of Azure nodepools, with YAML output",
			np: newNodePoolCollection(
				*newAzureNodePoolOT("1sad2", "s921a", "2021-01-02T15:04:32Z", "13.0.0", "test nodepool 1", 1, 3, -1, -1),
				*newAzureNodePoolOT("2a03f", "3a0d1", "2021-01-02T15:04:32Z", "13.0.0", "test nodepool 2", 3, 10, -1, -1),
				*newAzureNodePoolOT("asd29", "s0a10", "2021-01-02T15:04:32Z", "13.2.0", "test nodepool 3", 10, 10, 10, 10),
				*newAzureNodePoolOT("f930q", "s921a", "2021-01-02T15:04:32Z", "13.0.0", "test nodepool 4", 3, 3, -1, -1),
				*newAzureNodePoolOT("9f012", "29sa0", "2021-01-02T15:04:32Z", "13.2.0", "test nodepool 5", 0, 3, 1, 1),
				*newAzureNodePoolOT("2f0as", "s00sn", "2021-01-02T15:04:32Z", "13.1.0", "test nodepool 6", 2, 5, -1, -1),
			),
			provider:           key.ProviderAzure,
			outputType:         output.TypeYAML,
			expectedGoldenFile: "print_list_of_azure_nodepools_yaml_output.golden",
		},
		{
			name: "case 11: print list of Azure nodepools, with name output",
			np: newNodePoolCollection(
				*newAzureNodePoolOT("1sad2", "s921a", "2021-01-02T15:04:32Z", "13.0.0", "test nodepool 1", 1, 3, -1, -1),
				*newAzureNodePoolOT("2a03f", "3a0d1", "2021-01-02T15:04:32Z", "13.0.0", "test nodepool 2", 3, 10, -1, -1),
				*newAzureNodePoolOT("asd29", "s0a10", "2021-01-02T15:04:32Z", "13.2.0", "test nodepool 3", 10, 10, 10, 10),
				*newAzureNodePoolOT("f930q", "s921a", "2021-01-02T15:04:32Z", "13.0.0", "test nodepool 4", 3, 3, -1, -1),
				*newAzureNodePoolOT("9f012", "29sa0", "2021-01-02T15:04:32Z", "13.2.0", "test nodepool 5", 0, 3, 1, 1),
				*newAzureNodePoolOT("2f0as", "s00sn", "2021-01-02T15:04:32Z", "13.1.0", "test nodepool 6", 2, 5, -1, -1),
			),
			provider:           key.ProviderAzure,
			outputType:         output.TypeName,
			expectedGoldenFile: "print_list_of_azure_nodepools_name_output.golden",
		},
		{
			name:               "case 12: print single Azure nodepool, with table output",
			np:                 newAzureNodePoolOT("f930q", "s921a", time.Now().Format(time.RFC3339), "13.0.0", "test nodepool 4", 3, 3, -1, -1),
			provider:           key.ProviderAzure,
			outputType:         output.TypeDefault,
			expectedGoldenFile: "print_single_azure_nodepool_table_output.golden",
		},
		{
			name:               "case 13: print single Azure nodepool, with JSON output",
			np:                 newAzureNodePoolOT("f930q", "s921a", "2021-01-02T15:04:32Z", "13.0.0", "test nodepool 4", 3, 3, -1, -1),
			provider:           key.ProviderAzure,
			outputType:         output.TypeJSON,
			expectedGoldenFile: "print_single_azure_nodepool_json_output.golden",
		},
		{
			name:               "case 14: print single Azure nodepool, with YAML output",
			np:                 newAzureNodePoolOT("f930q", "s921a", "2021-01-02T15:04:32Z", "13.0.0", "test nodepool 4", 3, 3, -1, -1),
			provider:           key.ProviderAzure,
			outputType:         output.TypeYAML,
			expectedGoldenFile: "print_single_azure_nodepool_yaml_output.golden",
		},
		{
			name:               "case 15: print single Azure nodepool, with name output",
			np:                 newAzureNodePoolOT("f930q", "s921a", "2021-01-02T15:04:32Z", "13.2.0", "test nodepool 4", 3, 3, -1, -1),
			provider:           key.ProviderAzure,
			outputType:         output.TypeName,
			expectedGoldenFile: "print_single_azure_nodepool_name_output.golden",
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

func newAWSMachineDeploymentOT(name, clusterName, created, release, description string, nodesMin, nodesMax int) *infrastructurev1alpha3.AWSMachineDeployment {
	return newAWSMachineDeployment(name, clusterName, release, description, parseCreated(created), nodesMin, nodesMax)
}

func newcapiMachineDeploymentOT(name, clusterName, created, release string, nodesDesired, nodesReady int) *capi.MachineDeployment {
	return newcapiMachineDeployment(name, clusterName, release, parseCreated(created), nodesDesired, nodesReady)
}

func newAWSNodePoolOT(name, clusterName, created, release, description string, nodesMin, nodesMax, nodesDesired, nodesReady int) *nodepool.Nodepool {
	return newAWSNodePool(name, clusterName, release, description, parseCreated(created), nodesMin, nodesMax, nodesDesired, nodesReady)
}

func newAzureMachinePoolOT(name, clusterName, created, release string) *capzexp.AzureMachinePool {
	return newAzureMachinePool(name, clusterName, release, parseCreated(created))
}

func newCAPIexpMachinePoolOT(name, clusterName, created, release, description string, nodesDesired, nodesReady, nodesMin, nodesMax int) *capiexp.MachinePool {
	return newCAPIexpMachinePool(name, clusterName, release, description, parseCreated(created), nodesDesired, nodesReady, nodesMin, nodesMax)
}

func newAzureNodePoolOT(name, clusterName, created, release, description string, nodesMin, nodesMax, nodesDesired, nodesReady int) *nodepool.Nodepool {
	return newAzureNodePool(name, clusterName, release, description, parseCreated(created), nodesMin, nodesMax, nodesDesired, nodesReady)
}
