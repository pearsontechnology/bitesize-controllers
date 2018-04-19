/*
Copyright 2018 Pearson Technology

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
package v1

import (
	v1 "github.com/pearsontechnology/bitesize-controllers/vault-controller/pkg/apis/vault.local/v1"
	scheme "github.com/pearsontechnology/bitesize-controllers/vault-controller/pkg/client/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// VaultPoliciesGetter has a method to return a VaultPolicyInterface.
// A group's client should implement this interface.
type VaultPoliciesGetter interface {
	VaultPolicies() VaultPolicyInterface
}

// VaultPolicyInterface has methods to work with VaultPolicy resources.
type VaultPolicyInterface interface {
	Create(*v1.VaultPolicy) (*v1.VaultPolicy, error)
	Update(*v1.VaultPolicy) (*v1.VaultPolicy, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.VaultPolicy, error)
	List(opts meta_v1.ListOptions) (*v1.VaultPolicyList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.VaultPolicy, err error)
	VaultPolicyExpansion
}

// vaultpolicies implements VaultPolicyInterface
type vaultpolicies struct {
	client rest.Interface
}

// newVaultPolicies returns a VaultPolicies
func newVaultPolicies(c *VaultPolicyV1Client) *vaultpolicies {
	return &vaultpolicies{
		client: c.RESTClient(),
	}
}

// Get takes name of the vaultpolicy, and returns the corresponding vaultpolicy object, and an error if there is any.
func (c *vaultpolicies) Get(name string, options meta_v1.GetOptions) (result *v1.VaultPolicy, err error) {
	result = &v1.VaultPolicy{}
	err = c.client.Get().
		Resource("vaultpolicies").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VaultPolicies that match those selectors.
func (c *vaultpolicies) List(opts meta_v1.ListOptions) (result *v1.VaultPolicyList, err error) {
	result = &v1.VaultPolicyList{}
	err = c.client.Get().
		Resource("vaultpolicies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested vaultpolicies.
func (c *vaultpolicies) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("vaultpolicies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a vaultpolicy and creates it.  Returns the server's representation of the vaultpolicy, and an error, if there is any.
func (c *vaultpolicies) Create(vaultpolicy *v1.VaultPolicy) (result *v1.VaultPolicy, err error) {
	result = &v1.VaultPolicy{}
	err = c.client.Post().
		Resource("vaultpolicies").
		Body(vaultpolicy).
		Do().
		Into(result)
	return
}

// Update takes the representation of a vaultpolicy and updates it. Returns the server's representation of the vaultpolicy, and an error, if there is any.
func (c *vaultpolicies) Update(vaultpolicy *v1.VaultPolicy) (result *v1.VaultPolicy, err error) {
	result = &v1.VaultPolicy{}
	err = c.client.Put().
		Resource("vaultpolicies").
		Name(vaultpolicy.Name).
		Body(vaultpolicy).
		Do().
		Into(result)
	return
}

// Delete takes name of the vaultpolicy and deletes it. Returns an error if one occurs.
func (c *vaultpolicies) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("vaultpolicies").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *vaultpolicies) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Resource("vaultpolicies").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched vaultpolicy.
func (c *vaultpolicies) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.VaultPolicy, err error) {
	result = &v1.VaultPolicy{}
	err = c.client.Patch(pt).
		Resource("vaultpolicies").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
