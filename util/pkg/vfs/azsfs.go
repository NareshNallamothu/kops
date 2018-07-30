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

package vfs

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2018-02-01/storage"
	blob "github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
	"github.com/golang/glog"
	"k8s.io/kops/util/pkg/hashing"
)

var (
	blobFormatString = "https://%s.blob.core.windows.net"
	containerName    = "kops"
)

type AZSPath struct {
	client         *storage.AccountsClient
	resourceGroup  string
	storageAccount string
	key            string
	md5Hash        string
	accountKey     string
}

var _ Path = &AZSPath{}
var _ HasHash = &AZSPath{}

// AZSAcl is an ACL implementation for objects on Azure storage
type AZSAcl struct {
	RequestACL *string
}

func NewAZSPath(client *storage.AccountsClient, resourceGroup string, storageAccount string, key string) *AZSPath {
	resourceGroup = strings.TrimSuffix(resourceGroup, "/")
	storageAccount = strings.TrimPrefix(storageAccount, "/")
	key = strings.TrimPrefix(key, "/")

	return &AZSPath{
		client:         client,
		resourceGroup:  resourceGroup,
		storageAccount: storageAccount,
		key:            key,
	}
}

func (p *AZSPath) Path() string {
	return "azs://" + p.resourceGroup + "/" + p.storageAccount + "/" + p.key
}

func (p *AZSPath) Bucket() string {
	return p.storageAccount
}

func (p *AZSPath) Object() string {
	return p.storageAccount
}

func (p *AZSPath) Client() *storage.AccountsClient {
	return p.client
}

func (p *AZSPath) String() string {
	return p.Path()
}

func (p *AZSPath) Remove() error {
	ctx := context.Background()
	accountKey, accountKeyErr := p.getAccountKey(ctx)
	if accountKeyErr != nil {
		return accountKeyErr
	}

	blobURL := GetBlockBlobURL(ctx, accountKey, p.resourceGroup, p.storageAccount, containerName, p.key)

	_, err := blobURL.Delete(ctx, blob.DeleteSnapshotsOptionNone, blob.BlobAccessConditions{})
	if err != nil {
		return err
	}

	return nil
}

func (p *AZSPath) Join(relativePath ...string) Path {
	args := []string{p.key}
	args = append(args, relativePath...)
	joined := path.Join(args...)

	return &AZSPath{
		client:         p.client,
		resourceGroup:  p.resourceGroup,
		storageAccount: p.storageAccount,
		key:            joined,
	}
}

func (p *AZSPath) WriteFile(data io.ReadSeeker, acl ACL) error {
	ctx := context.Background()
	accountKey, err := p.getAccountKey(ctx)
	if err != nil {
		return err
	}

	blobURL := GetBlockBlobURL(ctx, accountKey, p.resourceGroup, p.storageAccount, containerName, p.key)

	glog.V(4).Infof("Writing file %q", p)

	// TODO: BP Encryption????
	// TODO: BP ACL?
	_, putErr := blobURL.PutBlob(
		ctx,
		data,
		blob.BlobHTTPHeaders{
			ContentType: "text/plain",
		},
		blob.Metadata{},
		blob.BlobAccessConditions{},
	)

	if putErr != nil {
		return putErr
	}

	return nil
}

// To prevent concurrent creates on the same file while maintaining atomicity of writes,
// we take a process-wide lock during the operation.
// Not a great approach, but fine for a single process (with low concurrency)
// TODO: should we enable versioning?
var createFileLockAZS sync.Mutex

func (p *AZSPath) CreateFile(data io.ReadSeeker, acl ACL) error {
	createFileLockAZS.Lock()
	defer createFileLockAZS.Unlock()

	// Check if exists
	_, err := p.ReadFile()
	if err == nil {
		return os.ErrExist
	}

	if !os.IsNotExist(err) {
		return err
	}

	return p.WriteFile(data, acl)
}

// ReadFile implements Path::ReadFile
func (p *AZSPath) ReadFile() ([]byte, error) {
	var b bytes.Buffer
	_, err := p.WriteTo(&b)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (p *AZSPath) WriteTo(out io.Writer) (int64, error) {

	ctx := context.Background()
	accountKey, err := p.getAccountKey(ctx)
	if err != nil {
		return 0, err
	}

	glog.V(4).Infof("Reading file %q", p)

	blobURL := GetBlockBlobURL(ctx, accountKey, p.resourceGroup, p.storageAccount, containerName, p.key)

	response, err := blobURL.GetBlob(ctx, blob.BlobRange{}, blob.BlobAccessConditions{}, false)
	if err != nil {
		if storageError, ok := GetStorageError(err); ok {
			switch storageError.ServiceCode() {
			case blob.ServiceCodeContainerNotFound:
				createErr := CreateContainerIfNotExists(ctx, accountKey, p.resourceGroup, p.storageAccount, containerName)
				if createErr != nil {
					return 0, fmt.Errorf("Error creating storage container: %v", createErr)
				}

				return 0, os.ErrNotExist
			case blob.ServiceCodeBlobNotFound:
				return 0, os.ErrNotExist
			default:
				return 0, fmt.Errorf("error fetching %s: %v", p, err)
			}
		}

		return 0, fmt.Errorf("error fetching %s: %v", p, err)
	}
	defer response.Body().Close()

	n, err := io.Copy(out, response.Body())
	if err != nil {
		return n, fmt.Errorf("error reading %s: %v", p, err)
	}
	return n, nil
}

func (p *AZSPath) ReadDir() ([]Path, error) {
	ctx := context.Background()
	accountKey, err := p.getAccountKey(ctx)
	if err != nil {
		return nil, err
	}

	prefix := p.key
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	var paths []Path
	c := GetContainerURL(ctx, accountKey, p.resourceGroup, p.storageAccount, containerName)
	blobsList, listErr := c.ListBlobs(
		ctx,
		blob.Marker{},
		blob.ListBlobsOptions{
			Prefix: prefix,
			Details: blob.BlobListingDetails{
				Snapshots: true,
			},
		})

	if listErr != nil {
		return nil, fmt.Errorf("error listing %s: %v", p, err)
	}

	for _, blob := range blobsList.Blobs.Blob {
		key := blob.Name
		child := &AZSPath{
			client:         p.client,
			resourceGroup:  p.resourceGroup,
			storageAccount: p.storageAccount,
			key:            key,
		}
		paths = append(paths, child)
	}

	glog.V(8).Infof("Listed files in %v: %v", p, paths)
	return paths, nil
}

func (p *AZSPath) ReadTree() ([]Path, error) {
	// TODO: BP Implement this
	return nil, fmt.Errorf("Not implemented")
}

func (p *AZSPath) Base() string {
	return path.Base(p.key)
}

func (p *AZSPath) PreferredHash() (*hashing.Hash, error) {
	return p.Hash(hashing.HashAlgorithmMD5)
}

func (p *AZSPath) Hash(a hashing.HashAlgorithm) (*hashing.Hash, error) {
	if a != hashing.HashAlgorithmMD5 {
		return nil, nil
	}

	md5 := p.md5Hash
	if md5 == "" {
		return nil, nil
	}

	md5Bytes, err := hex.DecodeString(md5)
	if err != nil {
		return nil, fmt.Errorf("Etag was not a valid MD5 sum: %q", md5)
	}

	return &hashing.Hash{Algorithm: hashing.HashAlgorithmMD5, HashValue: md5Bytes}, nil
}

func (p *AZSPath) getAccountKey(ctx context.Context) (string, error) {

	if p.accountKey != "" {
		return p.accountKey, nil
	}

	res, err := p.client.ListKeys(ctx, p.resourceGroup, p.storageAccount)

	if err != nil {
		return "", err
	}

	p.accountKey = *(((*res.Keys)[0]).Value)

	return p.accountKey, nil
}

func GetBlockBlobURL(ctx context.Context, accountKey, resourceGroupName, accountName, containerName, blobName string) blob.BlockBlobURL {
	container := GetContainerURL(ctx, accountKey, resourceGroupName, accountName, containerName)
	blob := container.NewBlockBlobURL(blobName)
	return blob
}

func GetContainerURL(ctx context.Context, accountKey, resourceGroupName, accountName, containerName string) blob.ContainerURL {
	c := blob.NewSharedKeyCredential(accountName, accountKey)
	pipeline := blob.NewPipeline(c, blob.PipelineOptions{
		Telemetry: blob.TelemetryOptions{Value: "kops"},
	})
	u, _ := url.Parse(fmt.Sprintf(blobFormatString, accountName))
	service := blob.NewServiceURL(*u, pipeline)
	container := service.NewContainerURL(containerName)
	return container
}

func CreateContainerIfNotExists(ctx context.Context, accountKey, resourceGroupName, accountName, containerName string) error {

	containerURL := GetContainerURL(ctx, accountKey, resourceGroupName, accountName, containerName)
	_, err := containerURL.GetPropertiesAndMetadata(ctx, blob.LeaseAccessConditions{})

	if err != nil {
		if storageError, ok := GetStorageError(err); ok {
			switch storageError.ServiceCode() {
			case blob.ServiceCodeContainerNotFound:
				// Create container
				createErr := CreateContainer(ctx, accountKey, resourceGroupName, accountName, containerName)

				if createErr != nil {
					return fmt.Errorf("error creating container %s: %v", containerName, createErr)
				}
			default:
				return fmt.Errorf("error creating container %s: %v", containerName, err)
			}
		}
	}

	return nil
}

func CreateContainer(ctx context.Context, accountKey, resourceGroupName, accountName, containerName string) error {

	containerURL := GetContainerURL(ctx, accountKey, resourceGroupName, accountName, containerName)

	_, err := containerURL.Create(
		ctx,
		blob.Metadata{},
		blob.PublicAccessNone)

	if err != nil {
		return err
	}

	return nil
}

func GetStorageError(err error) (blob.StorageError, bool) {
	if azureError, ok := err.(blob.StorageError); ok {
		return azureError, true

	}

	return nil, false
}
