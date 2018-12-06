package azuremodel

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

type ResourceGroupBuilder struct {
	*AzureModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &ResourceGroupBuilder{}

func (b *ResourceGroupBuilder) Build(c *fi.ModelBuilderContext) error {

	t := &azuretasks.ResourceGroup{
		Name:      s(b.Cluster.Spec.ResourceGroup),
		Lifecycle: b.Lifecycle,
	}

	c.AddTask(t)

	return nil
}
