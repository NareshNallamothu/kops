package azure

import (
	"fmt"

	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

// TODO: BP Implement this
func ListResourcesAzure(azureCloud azure.AzureCloud, clusterName string, region string) (map[string]*resources.Resource, error) {
	return nil, fmt.Errorf("ListResourcesAzure: Not implemented")
}
