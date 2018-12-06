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

	disks "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	compute "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-06-01/compute"
	resources "github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
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
	ResourceGroup() *resources.GroupsClient
	Compute() *compute.VirtualMachinesClient
	Storage() *storage.AccountsClient
	Disk() *disks.DisksClient

	Region() string
	ResourceGroupName() string

	// DefaultInstanceType determines a suitable instance type for the specified instance group
	DefaultInstanceType(cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error)
}

type azureCloudImplementation struct {
	resourceGroup *resources.GroupsClient
	compute       *compute.VirtualMachinesClient
	storage       *storage.AccountsClient
	disk          *disks.DisksClient

	region            string
	resourceGroupName string
}

var _ fi.Cloud = &azureCloudImplementation{}

func NewAzureCloud(region string, resourceGroupName string, tags map[string]string) (AzureCloud, error) {
	c := &azureCloudImplementation{
		region:            region,
		resourceGroupName: resourceGroupName,
	}

	// TODO: BP Make this a const
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		return nil, fmt.Errorf("Error building Azure API client. AZURE_SUBSCRIPTION_ID env variable not found")
	}

	resourceGroupClient := resources.NewGroupsClient(subscriptionID)
	computeClient := compute.NewVirtualMachinesClient(subscriptionID)
	storageClient := storage.NewAccountsClient(subscriptionID)
	disksClient := disks.NewDisksClient(subscriptionID)

	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("Error building Azure API client: %v", err)
	}

	resourceGroupClient.Authorizer = authorizer
	computeClient.Authorizer = authorizer
	storageClient.Authorizer = authorizer
	disksClient.Authorizer = authorizer

	c.resourceGroup = &resourceGroupClient
	c.compute = &computeClient
	c.storage = &storageClient
	c.disk = &disksClient

	return c, nil
}

// Compute returns private struct element resourceGroup.
func (c *azureCloudImplementation) ResourceGroup() *resources.GroupsClient {
	return c.resourceGroup
}

// Compute returns private struct element compute.
func (c *azureCloudImplementation) Compute() *compute.VirtualMachinesClient {
	return c.compute
}

// Storage returns private struct element storage.
func (c *azureCloudImplementation) Storage() *storage.AccountsClient {
	return c.storage
}

// Disk returns private struct element disk.
func (c *azureCloudImplementation) Disk() *disks.DisksClient {
	return c.disk
}

func (c *azureCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderAzure
}

// Region returns private struct element region.
func (c *azureCloudImplementation) Region() string {
	return c.region
}

func (c *azureCloudImplementation) ResourceGroupName() string {
	return c.resourceGroupName
}

// DefaultInstanceType determines an instance type for the specified cluster & instance group
func (c *azureCloudImplementation) DefaultInstanceType(cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error) {
	switch ig.Spec.Role {
	case kops.InstanceGroupRoleMaster:
		return "Standard_D2_v2", nil

	case kops.InstanceGroupRoleNode:
		return "Standard_D2_v2", nil

	case kops.InstanceGroupRoleBastion:
		return "Standard_B1s", nil

	default:
		return "", fmt.Errorf("unhandled role %q", ig.Spec.Role)
	}
}

func (c *azureCloudImplementation) DNS() (dnsprovider.Interface, error) {
	return nil, fmt.Errorf("DNS() not implemented")
}

func (c *azureCloudImplementation) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return fmt.Errorf("DeleteGroup() not implemented")
}

func (c *azureCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	return fmt.Errorf("DeleteInstance() not implemented")
}

func (c *azureCloudImplementation) FindVPCInfo(vpcID string) (*fi.VPCInfo, error) {
	glog.Warningf("FindVPCInfo not (yet) implemented on Azure")
	return nil, nil
}

func (c *azureCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return nil, fmt.Errorf("GetCloudGroups() not implemented")
}
