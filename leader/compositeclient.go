package leader

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type CompositeClient struct {
	getter  crclient.Client
	creator crclient.Client
}

func NewCompositeClient(getter, creator crclient.Client) CompositeClient {
	return CompositeClient{getter, creator}
}

//
// needs to implement the Client interface
// https://godoc.org/sigs.k8s.io/controller-runtime/pkg/client#Client
//
// Client interface is composed of 3 other interfaces:
// Reader - https://godoc.org/sigs.k8s.io/controller-runtime/pkg/client#Reader
// Writer - https://godoc.org/sigs.k8s.io/controller-runtime/pkg/client#Writer
// StatusClient - https://godoc.org/sigs.k8s.io/controller-runtime/pkg/client#StatusClient
//
// This means you need to implement ALL of the methods of all 3 of the above
// interfaces, that will be 8 methods.
//
// You need to decide which client each method should call. I think the Reader
// methods should pass through to getter. The Writer methods to createor.
//

//
// Here's an example of the Get method for the Reader interface. This is one of
// the 8 methods that need to get implemented:
//
func (c CompositeClient) Get(ctx context.Context, key crclient.ObjectKey, obj runtime.Object) error {
	return c.getter.Get(ctx, key, obj)
}

func (c CompositeClient) List(ctx context.Context, list runtime.Object, opts ...crclient.ListOption) error {
	return c.getter.List(ctx, list, opts...)
}

func (c CompositeClient) Create(ctx context.Context, obj runtime.Object, opts ...crclient.CreateOption) error {
	return c.creator.Create(ctx, obj, opts...)
}

func (c CompositeClient) Delete(ctx context.Context, obj runtime.Object, opts ...crclient.DeleteOption) error {
	return c.creator.Delete(ctx, obj, opts...)
}

func (c CompositeClient) Update(ctx context.Context, obj runtime.Object, opts ...crclient.UpdateOption) error {
	return c.creator.Update(ctx, obj, opts...)
}

func (c CompositeClient) Patch(ctx context.Context, obj runtime.Object, patch crclient.Patch, opts ...crclient.PatchOption) error {
	return c.creator.Patch(ctx, obj, patch, opts...)
}

func (c CompositeClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...crclient.DeleteAllOfOption) error {
	return c.DeleteAllOf(ctx, obj, opts...)
}

func (c CompositeClient) Status() crclient.StatusWriter {
	return c.getter.Status()
}
