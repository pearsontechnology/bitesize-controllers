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
package fake

import (
	vaultpolicy_v1 "github.com/pearsontechnology/bitesize-controllers/vault-controller/pkg/apis/vault.local/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVaultPolicies implements VaultPolicyInterface
type FakeVaultPolicies struct {
	Fake *FakeVaultPolicyV1
}

var vaultpoliciesResource = schema.GroupVersionResource{Group: "vaultpolicy", Version: "v1", Resource: "policies"}

var vaultpoliciesKind = schema.GroupVersionKind{Group: "vaultpolicy", Version: "v1", Kind: "Policy"}

// Get takes name of the vaultpolicy, and returns the corresponding vaultpolicy object, and an error if there is any.
func (c *FakeVaultPolicies) Get(name string, options v1.GetOptions) (result *vaultpolicy_v1.VaultPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(vaultpoliciesResource, name), &vaultpolicy_v1.VaultPolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*vaultpolicy_v1.VaultPolicy), err
}

// List takes label and field selectors, and returns the list of VaultPolicies that match those selectors.
func (c *FakeVaultPolicies) List(opts v1.ListOptions) (result *vaultpolicy_v1.VaultPolicyList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(vaultpoliciesResource, vaultpoliciesKind, opts), &vaultpolicy_v1.VaultPolicyList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &vaultpolicy_v1.VaultPolicyList{}
	for _, item := range obj.(*vaultpolicy_v1.VaultPolicyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested vaultpolicies.
func (c *FakeVaultPolicies) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(vaultpoliciesResource, opts))
}

// Create takes the representation of a vaultpolicy and creates it.  Returns the server's representation of the vaultpolicy, and an error, if there is any.
func (c *FakeVaultPolicies) Create(vaultpolicy *vaultpolicy_v1.VaultPolicy) (result *vaultpolicy_v1.VaultPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(vaultpoliciesResource, vaultpolicy), &vaultpolicy_v1.VaultPolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*vaultpolicy_v1.VaultPolicy), err
}

// Update takes the representation of a vaultpolicy and updates it. Returns the server's representation of the vaultpolicy, and an error, if there is any.
func (c *FakeVaultPolicies) Update(vaultpolicy *vaultpolicy_v1.VaultPolicy) (result *vaultpolicy_v1.VaultPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(vaultpoliciesResource, vaultpolicy), &vaultpolicy_v1.VaultPolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*vaultpolicy_v1.VaultPolicy), err
}

// Delete takes name of the vaultpolicy and deletes it. Returns an error if one occurs.
func (c *FakeVaultPolicies) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(vaultpoliciesResource, name), &vaultpolicy_v1.VaultPolicy{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVaultPolicies) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(vaultpoliciesResource, listOptions)

	_, err := c.Fake.Invokes(action, &vaultpolicy_v1.VaultPolicyList{})
	return err
}

// Patch applies the patch and returns the patched vaultpolicy.
func (c *FakeVaultPolicies) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *vaultpolicy_v1.VaultPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(vaultpoliciesResource, name, data, subresources...), &vaultpolicy_v1.VaultPolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*vaultpolicy_v1.VaultPolicy), err
}
