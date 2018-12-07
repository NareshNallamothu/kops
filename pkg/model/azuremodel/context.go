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

package azuremodel

import (
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

type AzureModelContext struct {
	*model.KopsModelContext
}

// SafeObjectName returns the object name and cluster name escaped for Azure
func (c *AzureModelContext) SafeObjectName(name string) string {
	return azure.SafeObjectName(name, c.Cluster.ObjectMeta.Name)
}
