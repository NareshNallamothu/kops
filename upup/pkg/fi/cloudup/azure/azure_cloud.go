/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azure

import (
	"fmt"
	"os"

	compute "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-06-01/compute"
	storage "github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2018-02-01/storage"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

type AzureCloud interface {
	fi.Cloud

	// TODO: BP Incomplete
	Compute() *compute.VirtualMachinesClient
	Storage() *storage.AccountsClient
}

type azureCloudImplementation struct {
	compute *compute.VirtualMachinesClient
	storage *storage.AccountsClient
}

var _ fi.Cloud = &azureCloudImplementation{}

func NewAzureCloud() (AzureCloud, error) {
	c := &azureCloudImplementation{}

	// TODO: BP Make this a const
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		return nil, fmt.Errorf("Error building Azure API client. AZURE_SUBSCRIPTION_ID env variable not found")
	}

	computeClient := compute.NewVirtualMachinesClient(subscriptionID)
	storageClient := storage.NewAccountsClient(subscriptionID)

	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("Error building Azure API client: %v", err)
	}

	computeClient.Authorizer = authorizer
	storageClient.Authorizer = authorizer

	c.compute = &computeClient
	c.storage = &storageClient

	return c, nil
}

// Compute returns private struct element compute.
func (c *azureCloudImplementation) Compute() *compute.VirtualMachinesClient {
	return c.compute
}

// Storage returns private struct element storage.
func (c *azureCloudImplementation) Storage() *storage.AccountsClient {
	return c.storage
}

func (c *azureCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderAzure
}

func (c *azureCloudImplementation) DNS() (dnsprovider.Interface, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *azureCloudImplementation) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return fmt.Errorf("not implemented")
}

func (c *azureCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	return fmt.Errorf("not implemented")
}

func (c *azureCloudImplementation) FindVPCInfo(vpcID string) (*fi.VPCInfo, error) {
	glog.Warningf("FindVPCInfo not (yet) implemented on Azure")
	return nil, nil
}

func (c *azureCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return nil, fmt.Errorf("not implemented")
}
