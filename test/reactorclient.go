package test

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/testing"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type ReactorClient struct {
	testing.Fake
	client crclient.Client
}

func NewReactorClient(client crclient.Client) ReactorClient {
	return ReactorClient{client: client}
}

func (c ReactorClient) Get(ctx context.Context, key crclient.ObjectKey, obj runtime.Object) error {
	fmt.Println("XXX Entered Get")
	fmt.Printf("Group: %v\n", obj.GetObjectKind().GroupVersionKind().Group)
	fmt.Printf("Version: %v\n", obj.GetObjectKind().GroupVersionKind().Version)
	fmt.Printf("Resource: %v\n", obj.GetObjectKind().GroupVersionKind().Kind)
	resource, err := getGVRFromObject(obj, scheme.Scheme)
	if err != nil {
		return err
	}
	// resource := schema.GroupVersionResource{
	//     // Group:    obj.GetObjectKind().GroupVersionKind().Group,
	//     // Version:  obj.GetObjectKind().GroupVersionKind().Version,
	//     // Resource: obj.GetObjectKind().GroupVersionKind().Kind,
	//     Group:    "",
	//     Version:  "v1",
	//     Resource: "pod",
	// }
	fmt.Printf("XXX resource: %v\n", resource)

	actionCopy := testing.NewGetAction(resource, key.Namespace, key.Name)
	fmt.Printf("XXX action verb [%v]\n", actionCopy.GetVerb())
	fmt.Printf("XXX action resource [%v]\n", actionCopy.GetResource().Resource)
	for _, reactor := range c.ReactionChain {
		if !reactor.Handles(actionCopy) {
			fmt.Println("XXX reactor does NOT handle the action")
			continue
		}
	}

	clienterr := c.client.Get(ctx, key, obj)
	if clienterr != nil {

	}
	retobj, err := c.Fake.Invokes(testing.NewGetAction(resource, key.Namespace, key.Name), obj)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Println("XXX deepcopy retobj")
	fmt.Printf("XXX retobj: %v\n", retobj)
	obj = retobj.DeepCopyObject()
	return nil
}

func (c ReactorClient) List(ctx context.Context, list runtime.Object, opts ...crclient.ListOption) error {
	return c.client.List(ctx, list, opts...)
}

// func (f *testing.Fake) Create(ctx context.Context, obj runtime.Object, opts ...crclient.CreateOption) error {
//     return nil
// }

func (c ReactorClient) Create(ctx context.Context, obj runtime.Object, opts ...crclient.CreateOption) error {
	return c.client.Create(ctx, obj, opts...)
}

func (c ReactorClient) Delete(ctx context.Context, obj runtime.Object, opts ...crclient.DeleteOption) error {
	return c.client.Delete(ctx, obj, opts...)
}

func (c ReactorClient) Update(ctx context.Context, obj runtime.Object, opts ...crclient.UpdateOption) error {
	return c.client.Update(ctx, obj, opts...)
}

func (c ReactorClient) Patch(ctx context.Context, obj runtime.Object, patch crclient.Patch, opts ...crclient.PatchOption) error {
	return c.client.Patch(ctx, obj, patch, opts...)
}

func (c ReactorClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...crclient.DeleteAllOfOption) error {
	return c.DeleteAllOf(ctx, obj, opts...)
}

func (c ReactorClient) Status() crclient.StatusWriter {
	return c.client.Status()
}

// Copied from controller-runtime fake client.
func getGVRFromObject(obj runtime.Object, scheme *runtime.Scheme) (schema.GroupVersionResource, error) {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	return gvr, nil
}
