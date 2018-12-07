package azuremodel

import (
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

type AutoscalingGroupModelBuilder struct {
	*AzureModelContext

	BootstrapScript *model.BootstrapScript
	Lifecycle       *fi.Lifecycle
}

var _ fi.ModelBuilder = &AutoscalingGroupModelBuilder{}

func (b *AutoscalingGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, ig := range b.InstanceGroups {
		name := b.SafeObjectName(ig.ObjectMeta.Name)

		platformFaultDomainCount := fi.Int32(2)
		platformUpdateDomainCount := fi.Int32(3)

		t := &azuretasks.AvailabilitySet{
			Name:                      s(name),
			PlatformFaultDomainCount:  platformFaultDomainCount,
			PlatformUpdateDomainCount: platformUpdateDomainCount,
		}

		c.AddTask(t)
	}
	return nil
}
