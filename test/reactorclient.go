package test

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/testing"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

const (
	maxNameLength          = 63
	randomLength           = 5
	maxGeneratedNameLength = maxNameLength - randomLength
)

type ReactorClient struct {
	testing.Fake
	client crclient.Client
}

func NewReactorClient(client crclient.Client) ReactorClient {
	return ReactorClient{client: client}
}

func (c ReactorClient) Get(ctx context.Context, key crclient.ObjectKey, obj runtime.Object) error {
	fmt.Println("YYY Entered Get")
	fmt.Printf("Group: %v\n", obj.GetObjectKind().GroupVersionKind().Group)
	fmt.Printf("Version: %v\n", obj.GetObjectKind().GroupVersionKind().Version)
	fmt.Printf("Resource: %v\n", obj.GetObjectKind().GroupVersionKind().Kind)
	resource, err := getGVRFromObject(obj, scheme.Scheme)
	if err != nil {
		return err
	}
	fmt.Printf("YYY resource: %v\n", resource)

	retobj, err := c.Fake.Invokes(testing.NewGetAction(resource, key.Namespace, key.Name), obj)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	if retobj == obj {
		fmt.Println("YYY invoke returned the default object")
		return c.client.Get(ctx, key, obj)
	}
	return nil
}

func (c ReactorClient) List(ctx context.Context, list runtime.Object, opts ...crclient.ListOption) error {
	gvk, err := apiutil.GVKForObject(list, scheme.Scheme)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(gvk.Kind, "List") {
		return fmt.Errorf("non-list type %T (kind %q) passed as output", list, gvk)
	}
	// we need the non-list GVK, so chop off the "List" from the end of the kind
	gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]

	resource, err := getGVRFromObject(list, scheme.Scheme)
	if err != nil {
		return err
	}

	listOpts := crclient.ListOptions{}
	listOpts.ApplyOptions(opts)

	retobj, err := c.Fake.Invokes(testing.NewListAction(resource, gvk,
		listOpts.Namespace, *listOpts.AsListOptions()), list)
	if err != nil {
		return err
	}
	if retobj == list {
		return c.client.List(ctx, list, opts...)
	}
	return nil
}

func (c ReactorClient) Create(ctx context.Context, obj runtime.Object, opts ...crclient.CreateOption) error {
	resource, err := getGVRFromObject(obj, scheme.Scheme)
	if err != nil {
		return err
	}

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	if accessor.GetName() == "" && accessor.GetGenerateName() != "" {
		base := accessor.GetGenerateName()
		if len(base) > maxGeneratedNameLength {
			base = base[:maxGeneratedNameLength]
		}
		accessor.SetName(fmt.Sprintf("%s%s", base, utilrand.String(randomLength)))
	}

	retobj, err := c.Fake.Invokes(testing.NewCreateAction(resource, accessor.GetNamespace(), obj), obj)
	if err != nil {
		fmt.Printf("Invoke failed to create: %v\n", err.Error())
		return err
	}

	if retobj == obj {
		fmt.Println("YYY invoke returned the default object")
		return c.client.Create(ctx, obj, opts...)
	}
	return nil
}

func (c ReactorClient) Delete(ctx context.Context, obj runtime.Object, opts ...crclient.DeleteOption) error {
	resource, err := getGVRFromObject(obj, scheme.Scheme)
	if err != nil {
		return err
	}

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	retobj, err := c.Fake.Invokes(testing.NewDeleteAction(resource, accessor.GetNamespace(), accessor.GetName()), obj)
	if err != nil {
		fmt.Printf("Invoke failed to delete: %v\n", err.Error())
		return err
	}
	if retobj == obj {
		fmt.Println("YYY invoke returned the default object")
		return c.client.Delete(ctx, obj, opts...)
	}
	return nil
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
