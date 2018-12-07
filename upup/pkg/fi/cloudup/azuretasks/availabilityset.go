package azuretasks

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-06-01/compute"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

type AvailabilitySet struct {
	Name                      *string
	Lifecycle                 *fi.Lifecycle
	Location                  *string
	PlatformFaultDomainCount  *int32
	PlatformUpdateDomainCount *int32
	Tags                      map[string]string
}

var _ fi.CompareWithID = &AvailabilitySet{}

func (e *AvailabilitySet) CompareWithID() *string {
	return e.Name
}

func (e *AvailabilitySet) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return GetResourceGroupDependency(tasks)
}

func (e *AvailabilitySet) Find(c *fi.Context) (*AvailabilitySet, error) {
	cloud := c.Cloud.(azure.AzureCloud)
	ctx := context.Background()

	r, err := cloud.AvailabilitySet().Get(ctx, cloud.ResourceGroupName(), *e.Name)
	if err != nil {
		if azure.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("Error finding Availability Set: %v", err)
	}

	actual := &AvailabilitySet{
		Name:                      r.Name,
		Lifecycle:                 e.Lifecycle,
		Location:                  r.Location,
		PlatformFaultDomainCount:  r.PlatformFaultDomainCount,
		PlatformUpdateDomainCount: r.PlatformUpdateDomainCount,
		Tags:                      fi.StringMapValue(r.Tags),
	}

	return actual, nil

}

func (e *AvailabilitySet) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *AvailabilitySet) CheckChanges(a, e, changes *AvailabilitySet) error {
	return nil
}

func (_ *AvailabilitySet) RenderAzure(t *azure.AzureAPITarget, a, e, changes *AvailabilitySet) error {
	cloud := t.Cloud
	ctx := context.Background()

	_, err := cloud.AvailabilitySet().CreateOrUpdate(ctx,
		cloud.ResourceGroupName(),
		*e.Name,
		compute.AvailabilitySet{
			Location: fi.String(cloud.Region()),
			AvailabilitySetProperties: &compute.AvailabilitySetProperties{
				PlatformFaultDomainCount:  e.PlatformFaultDomainCount,
				PlatformUpdateDomainCount: e.PlatformUpdateDomainCount,
			},
			Sku: &compute.Sku{
				Name: fi.String("Aligned"),
			},
		})

	if err != nil {
		return fmt.Errorf("Error creating Availability Set: %v", err)
	}

	return nil
}
