package provider

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/kubectl-gs/internal/feature"
	"github.com/giantswarm/kubectl-gs/internal/key"
	"github.com/giantswarm/kubectl-gs/pkg/data/domain/nodepool"
)

func GetAWSTable(npResource nodepool.Resource, capabilities *feature.Service) *metav1.Table {
	table := &metav1.Table{
		ColumnDefinitions: []metav1.TableColumnDefinition{
			{Name: "ID", Type: "string"},
			{Name: "Created", Type: "string", Format: "date-time"},
			{Name: "Condition", Type: "string"},
			{Name: "Nodes Min/Max", Type: "string"},
			{Name: "Nodes Desired", Type: "integer"},
			{Name: "Nodes Ready", Type: "integer"},
			{Name: "Description", Type: "string"},
		},
	}

	switch n := npResource.(type) {
	case *nodepool.Nodepool:
		table.Rows = append(table.Rows, getAWSNodePoolRow(*n, capabilities))
	case *nodepool.Collection:
		for _, nodePool := range n.Items {
			table.Rows = append(table.Rows, getAWSNodePoolRow(nodePool, capabilities))
		}
	}

	return table
}

func getAWSNodePoolRow(
	nodePool nodepool.Nodepool,
	capabilities *feature.Service,
) metav1.TableRow {
	if nodePool.MachineDeployment == nil || nodePool.AWSMachineDeployment == nil {
		return metav1.TableRow{}
	}

	return metav1.TableRow{
		Cells: []interface{}{
			nodePool.MachineDeployment.GetName(),
			nodePool.MachineDeployment.CreationTimestamp.UTC(),
			getAWSLatestCondition(nodePool, capabilities),
			getAWSAutoscaling(nodePool, capabilities),
			nodePool.MachineDeployment.Status.Replicas,
			nodePool.MachineDeployment.Status.ReadyReplicas,
			getAWSDescription(nodePool),
		},
		Object: runtime.RawExtension{
			Object: nodePool.MachineDeployment,
		},
	}
}

func getAWSLatestCondition(nodePool nodepool.Nodepool, capabilities *feature.Service) string {
	releaseVersion := key.ReleaseVersion(nodePool.MachineDeployment)
	isSupported := capabilities.Supports(feature.NodePoolConditions, releaseVersion)
	if !isSupported {
		return naValue
	}

	// Unsupported feature.

	return naValue
}

func getAWSAutoscaling(nodePool nodepool.Nodepool, capabilities *feature.Service) string {
	releaseVersion := key.ReleaseVersion(nodePool.MachineDeployment)
	isSupported := capabilities.Supports(feature.Autoscaling, releaseVersion)
	if !isSupported {
		return naValue
	}

	minScaling := nodePool.AWSMachineDeployment.Spec.NodePool.Scaling.Min
	maxScaling := nodePool.AWSMachineDeployment.Spec.NodePool.Scaling.Max

	return fmt.Sprintf("%d/%d", minScaling, maxScaling)
}

func getAWSDescription(nodePool nodepool.Nodepool) string {
	description := nodePool.AWSMachineDeployment.Spec.NodePool.Description
	if len(description) < 1 {
		description = naValue
	}

	return description
}
