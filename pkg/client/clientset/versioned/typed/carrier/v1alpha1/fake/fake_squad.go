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

package fake

import (
	"context"

	v1alpha1 "github.com/ocgi/carrier/pkg/apis/carrier/v1alpha1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSquads implements SquadInterface
type FakeSquads struct {
	Fake *FakeCarrierV1alpha1
	ns   string
}

var squadsResource = schema.GroupVersionResource{Group: "carrier.ocgi.dev", Version: "v1alpha1", Resource: "squads"}

var squadsKind = schema.GroupVersionKind{Group: "carrier.ocgi.dev", Version: "v1alpha1", Kind: "Squad"}

// Get takes name of the squad, and returns the corresponding squad object, and an error if there is any.
func (c *FakeSquads) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Squad, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(squadsResource, c.ns, name), &v1alpha1.Squad{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Squad), err
}

// List takes label and field selectors, and returns the list of Squads that match those selectors.
func (c *FakeSquads) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SquadList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(squadsResource, squadsKind, c.ns, opts), &v1alpha1.SquadList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SquadList{ListMeta: obj.(*v1alpha1.SquadList).ListMeta}
	for _, item := range obj.(*v1alpha1.SquadList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested squads.
func (c *FakeSquads) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(squadsResource, c.ns, opts))

}

// Create takes the representation of a squad and creates it.  Returns the server's representation of the squad, and an error, if there is any.
func (c *FakeSquads) Create(ctx context.Context, squad *v1alpha1.Squad, opts v1.CreateOptions) (result *v1alpha1.Squad, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(squadsResource, c.ns, squad), &v1alpha1.Squad{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Squad), err
}

// Update takes the representation of a squad and updates it. Returns the server's representation of the squad, and an error, if there is any.
func (c *FakeSquads) Update(ctx context.Context, squad *v1alpha1.Squad, opts v1.UpdateOptions) (result *v1alpha1.Squad, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(squadsResource, c.ns, squad), &v1alpha1.Squad{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Squad), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSquads) UpdateStatus(ctx context.Context, squad *v1alpha1.Squad, opts v1.UpdateOptions) (*v1alpha1.Squad, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(squadsResource, "status", c.ns, squad), &v1alpha1.Squad{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Squad), err
}

// Delete takes name of the squad and deletes it. Returns an error if one occurs.
func (c *FakeSquads) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(squadsResource, c.ns, name, opts), &v1alpha1.Squad{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSquads) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(squadsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.SquadList{})
	return err
}

// Patch applies the patch and returns the patched squad.
func (c *FakeSquads) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Squad, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(squadsResource, c.ns, name, pt, data, subresources...), &v1alpha1.Squad{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Squad), err
}

// GetScale takes name of the squad, and returns the corresponding scale object, and an error if there is any.
func (c *FakeSquads) GetScale(ctx context.Context, squadName string, options v1.GetOptions) (result *autoscalingv1.Scale, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetSubresourceAction(squadsResource, c.ns, "scale", squadName), &autoscalingv1.Scale{})

	if obj == nil {
		return nil, err
	}
	return obj.(*autoscalingv1.Scale), err
}

// UpdateScale takes the representation of a scale and updates it. Returns the server's representation of the scale, and an error, if there is any.
func (c *FakeSquads) UpdateScale(ctx context.Context, squadName string, scale *autoscalingv1.Scale, opts v1.UpdateOptions) (result *autoscalingv1.Scale, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(squadsResource, "scale", c.ns, scale), &autoscalingv1.Scale{})

	if obj == nil {
		return nil, err
	}
	return obj.(*autoscalingv1.Scale), err
}
