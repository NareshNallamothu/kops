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

package azuretasks

import (
	"fmt"

	"k8s.io/kops/upup/pkg/fi"
)

type Instance struct {
	ID        *string
	Name      *string
	Lifecycle *fi.Lifecycle
}

var _ fi.CompareWithID = &Instance{}

func (s *Instance) CompareWithID() *string {
	return s.ID
}

// TODO: BP Implement
func (e *Instance) Find(c *fi.Context) (*Instance, error) {
	// cloud := c.Cloud.(azure.AzureCloud)

	return nil, fmt.Errorf("azure instance.Find not implemented")
}

func (e *Instance) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

// TODO: BP Implement
func (_ *Instance) CheckChanges(a, e, changes *Instance) error {
	return fmt.Errorf("azure instance.CheckChanges not implemented")
}
