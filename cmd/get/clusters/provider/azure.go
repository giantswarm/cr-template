package provider

import (
	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/kubectl-gs/pkg/data/domain/cluster"
)

func GetAzureTable(resource runtime.Object) *metav1.Table {
	var clusterLists []runtime.Object
	{
		switch c := resource.(type) {
		case *cluster.CommonClusterList:
			clusterLists = c.Items
		default:
			clusterLists = []runtime.Object{resource}
		}
	}

	// Creating a custom table resource.
	table := &metav1.Table{}

	table.ColumnDefinitions = []metav1.TableColumnDefinition{
		{Name: "ID", Type: "string"},
		{Name: "Description", Type: "string"},
	}

	table.Rows = make([]metav1.TableRow, 0, len(clusterLists))
	for _, clusterList := range clusterLists {
		switch c := clusterList.(type) {
		case *corev1alpha1.AzureClusterConfigList:
			for _, currentCluster := range c.Items {
				table.Rows = append(table.Rows, getAzureClusterConfigRow(&currentCluster))
			}

		case *corev1alpha1.AzureClusterConfig:
			table.Rows = append(table.Rows, getAzureClusterConfigRow(c))

		default:
			continue
		}
	}

	return table
}

func getAzureClusterConfigRow(cr *corev1alpha1.AzureClusterConfig) metav1.TableRow {
	return metav1.TableRow{
		Cells: []interface{}{
			cr.Spec.Guest.ID,
			cr.Spec.Guest.Name,
		},
	}
}
