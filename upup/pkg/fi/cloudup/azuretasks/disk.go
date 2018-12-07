package azuretasks

import (
	"context"
	"fmt"

	disks "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

type Disk struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Location   *string
	VolumeType *string
	SizeGB     *int32
	Zone       *string
	Tags       map[string]string
}

var _ fi.CompareWithID = &Disk{}

func (e *Disk) CompareWithID() *string {
	return e.Name
}

func (e *Disk) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return GetResourceGroupDependency(tasks)
}

func (e *Disk) Find(c *fi.Context) (*Disk, error) {
	cloud := c.Cloud.(azure.AzureCloud)
	ctx := context.Background()

	r, err := cloud.Disk().Get(ctx, cloud.ResourceGroupName(), *e.Name)
	if err != nil {
		if azure.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("Error finding Disk: %v", err)
	}

	actual := &Disk{
		Name:      r.Name,
		Lifecycle: e.Lifecycle,
		Location:  r.Location,
		SizeGB:    r.DiskProperties.DiskSizeGB,
		Tags:      fi.StringMapValue(r.Tags),
	}

	return actual, nil
}

func (e *Disk) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Disk) CheckChanges(a, e, changes *Disk) error {
	if a != nil {
		if changes.SizeGB != nil {
			return fi.CannotChangeField("SizeGB")
		}
		if changes.Zone != nil {
			return fi.CannotChangeField("Zone")
		}
		if changes.VolumeType != nil {
			return fi.CannotChangeField("VolumeType")
		}
	}
	return nil
}

func (_ *Disk) RenderAzure(t *azure.AzureAPITarget, a, e, changes *Disk) error {
	cloud := t.Cloud
	ctx := context.Background()

	future, err := cloud.Disk().CreateOrUpdate(
		ctx,
		cloud.ResourceGroupName(),
		*e.Name,
		disks.Disk{
			Location: e.Location,
			Sku: &disks.DiskSku{
				Name: disks.StorageAccountTypesPremiumLRS,
			},
			DiskProperties: &disks.DiskProperties{
				CreationData: &disks.CreationData{
					CreateOption: disks.Empty,
				},
				DiskSizeGB: e.SizeGB,
			},
			Tags: fi.StringMap(e.Tags),
		},
	)
	if err != nil {
		return fmt.Errorf("Error creating Disk: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, cloud.Disk().Client)
	if err != nil {
		return fmt.Errorf("Error creating Disk: %v", err)
	}

	return nil
}
