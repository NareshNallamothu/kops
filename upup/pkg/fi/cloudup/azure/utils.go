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
	"github.com/Azure/go-autorest/autorest"
)

func IsNotFound(err error) bool {
	apiErr, ok := err.(autorest.DetailedError)
	if !ok {
		return false
	}

	return apiErr.Response.StatusCode == 404
}

// SafeObjectName returns the object name and cluster name escaped for Azure
func SafeObjectName(name string, clusterName string) string {
	azureName := name + "-" + clusterName

	return azureName
}
