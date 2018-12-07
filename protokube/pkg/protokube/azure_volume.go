package protokube

import (
	"fmt"
)

// AzureVolumes is the Volumes implementation for Azure
type AzureVolumes struct {
}

var _ Volumes = &AzureVolumes{}

// AttachVolume attaches the specified volume to this instance, returning the mountpoint & nil if successful
func (v *AzureVolumes) AttachVolume(volume *Volume) error {
	return fmt.Errorf("azure_volume AttachVolume not implemented")
}

// FindMountedVolume implements Volumes::FindMountedVolume
func (v *AzureVolumes) FindMountedVolume(volume *Volume) (string, error) {
	return "", fmt.Errorf("azure_volume FindMountedVolume not implemented")
}

// FindVolumes implements Volumes::FindVolumes
func (v *AzureVolumes) FindVolumes() ([]*Volume, error) {
	return nil, fmt.Errorf("azure_volume FindVolumes not implemented")

}
