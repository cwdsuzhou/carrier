// Copyright 2021 The OCGI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/ocgi/carrier/pkg/apis/carrier/v1alpha1"
	scheme "github.com/ocgi/carrier/pkg/client/clientset/versioned/scheme"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SquadsGetter has a method to return a SquadInterface.
// A group's client should implement this interface.
type SquadsGetter interface {
	Squads(namespace string) SquadInterface
}

// SquadInterface has methods to work with Squad resources.
type SquadInterface interface {
	Create(*v1alpha1.Squad) (*v1alpha1.Squad, error)
	Update(*v1alpha1.Squad) (*v1alpha1.Squad, error)
	UpdateStatus(*v1alpha1.Squad) (*v1alpha1.Squad, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Squad, error)
	List(opts v1.ListOptions) (*v1alpha1.SquadList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Squad, err error)
	GetScale(squadName string, options v1.GetOptions) (*v1beta1.Scale, error)
	UpdateScale(squadName string, scale *v1beta1.Scale) (*v1beta1.Scale, error)

	SquadExpansion
}

// squads implements SquadInterface
type squads struct {
	client rest.Interface
	ns     string
}

// newSquads returns a Squads
func newSquads(c *CarrierV1alpha1Client, namespace string) *squads {
	return &squads{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the squad, and returns the corresponding squad object, and an error if there is any.
func (c *squads) Get(name string, options v1.GetOptions) (result *v1alpha1.Squad, err error) {
	result = &v1alpha1.Squad{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("squads").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Squads that match those selectors.
func (c *squads) List(opts v1.ListOptions) (result *v1alpha1.SquadList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.SquadList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("squads").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested squads.
func (c *squads) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("squads").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a squad and creates it.  Returns the server's representation of the squad, and an error, if there is any.
func (c *squads) Create(squad *v1alpha1.Squad) (result *v1alpha1.Squad, err error) {
	result = &v1alpha1.Squad{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("squads").
		Body(squad).
		Do().
		Into(result)
	return
}

// Update takes the representation of a squad and updates it. Returns the server's representation of the squad, and an error, if there is any.
func (c *squads) Update(squad *v1alpha1.Squad) (result *v1alpha1.Squad, err error) {
	result = &v1alpha1.Squad{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("squads").
		Name(squad.Name).
		Body(squad).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *squads) UpdateStatus(squad *v1alpha1.Squad) (result *v1alpha1.Squad, err error) {
	result = &v1alpha1.Squad{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("squads").
		Name(squad.Name).
		SubResource("status").
		Body(squad).
		Do().
		Into(result)
	return
}

// Delete takes name of the squad and deletes it. Returns an error if one occurs.
func (c *squads) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("squads").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *squads) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("squads").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched squad.
func (c *squads) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Squad, err error) {
	result = &v1alpha1.Squad{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("squads").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}

// GetScale takes name of the squad, and returns the corresponding v1beta1.Scale object, and an error if there is any.
func (c *squads) GetScale(squadName string, options v1.GetOptions) (result *v1beta1.Scale, err error) {
	result = &v1beta1.Scale{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("squads").
		Name(squadName).
		SubResource("scale").
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// UpdateScale takes the top resource name and the representation of a scale and updates it. Returns the server's representation of the scale, and an error, if there is any.
func (c *squads) UpdateScale(squadName string, scale *v1beta1.Scale) (result *v1beta1.Scale, err error) {
	result = &v1beta1.Scale{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("squads").
		Name(squadName).
		SubResource("scale").
		Body(scale).
		Do().
		Into(result)
	return
}
