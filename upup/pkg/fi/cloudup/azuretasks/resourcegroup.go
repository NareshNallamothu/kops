package azuretasks

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

type ResourceGroup struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Tags map[string]string
}

var _ fi.CompareWithID = &ResourceGroup{}

func (e *ResourceGroup) CompareWithID() *string {
	return e.Name
}

func (e *ResourceGroup) Find(c *fi.Context) (*ResourceGroup, error) {
	cloud := c.Cloud.(azure.AzureCloud)

	ctx := context.Background()
	r, err := cloud.ResourceGroup().Get(ctx, *e.Name)
	if err != nil {
		if azure.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("Error finding Resource Group: %v", err)
	}

	actual := &ResourceGroup{
		Name:      r.Name,
		Lifecycle: e.Lifecycle,
		Tags:      fi.StringMapValue(r.Tags),
	}

	return actual, nil
}

func (e *ResourceGroup) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *ResourceGroup) CheckChanges(a, e, changes *ResourceGroup) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}

	return nil
}

func (_ *ResourceGroup) RenderAzure(t *azure.AzureAPITarget, a, e, changes *ResourceGroup) error {
	cloud := t.Cloud
	ctx := context.Background()

	group := resources.Group{
		Location: fi.String(cloud.Region()),
		Tags:     fi.StringMap(e.Tags),
	}

	_, err := cloud.ResourceGroup().CreateOrUpdate(
		ctx,
		*e.Name,
		group,
	)

	if err != nil {
		return fmt.Errorf("Error saving Resource Group: %v", err)
	}

	return nil
}
