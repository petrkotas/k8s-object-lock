/*
dcs
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	lockv1 "github.com/petrkotas/k8s-object-lock/pkg/api/lock/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeLocks implements LockInterface
type FakeLocks struct {
	Fake *FakeLocksV1
	ns   string
}

var locksResource = schema.GroupVersionResource{Group: "locks.kotas.tech", Version: "v1", Resource: "locks"}

var locksKind = schema.GroupVersionKind{Group: "locks.kotas.tech", Version: "v1", Kind: "Lock"}

// Get takes name of the lock, and returns the corresponding lock object, and an error if there is any.
func (c *FakeLocks) Get(ctx context.Context, name string, options v1.GetOptions) (result *lockv1.Lock, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(locksResource, c.ns, name), &lockv1.Lock{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lockv1.Lock), err
}

// List takes label and field selectors, and returns the list of Locks that match those selectors.
func (c *FakeLocks) List(ctx context.Context, opts v1.ListOptions) (result *lockv1.LockList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(locksResource, locksKind, c.ns, opts), &lockv1.LockList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &lockv1.LockList{ListMeta: obj.(*lockv1.LockList).ListMeta}
	for _, item := range obj.(*lockv1.LockList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested locks.
func (c *FakeLocks) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(locksResource, c.ns, opts))

}

// Create takes the representation of a lock and creates it.  Returns the server's representation of the lock, and an error, if there is any.
func (c *FakeLocks) Create(ctx context.Context, lock *lockv1.Lock, opts v1.CreateOptions) (result *lockv1.Lock, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(locksResource, c.ns, lock), &lockv1.Lock{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lockv1.Lock), err
}

// Update takes the representation of a lock and updates it. Returns the server's representation of the lock, and an error, if there is any.
func (c *FakeLocks) Update(ctx context.Context, lock *lockv1.Lock, opts v1.UpdateOptions) (result *lockv1.Lock, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(locksResource, c.ns, lock), &lockv1.Lock{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lockv1.Lock), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeLocks) UpdateStatus(ctx context.Context, lock *lockv1.Lock, opts v1.UpdateOptions) (*lockv1.Lock, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(locksResource, "status", c.ns, lock), &lockv1.Lock{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lockv1.Lock), err
}

// Delete takes name of the lock and deletes it. Returns an error if one occurs.
func (c *FakeLocks) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(locksResource, c.ns, name), &lockv1.Lock{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeLocks) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(locksResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &lockv1.LockList{})
	return err
}

// Patch applies the patch and returns the patched lock.
func (c *FakeLocks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *lockv1.Lock, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(locksResource, c.ns, name, pt, data, subresources...), &lockv1.Lock{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lockv1.Lock), err
}
