package crds

import (
	"context"

	"github.com/platfornow/lash/internal/core"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func PatchAndDelete(ctx context.Context, restConfig *rest.Config, el *unstructured.Unstructured) error {
	err := core.Patch(ctx, core.PatchOpts{
		RESTConfig: restConfig,
		GVK:        el.GroupVersionKind(),
		PatchData:  []byte(`{"metadata":{"finalizers":[]}}`),
		PatchType:  types.StrategicMergePatchType,
		Name:       el.GetName(),
		Namespace:  el.GetNamespace(),
	})
	if err != nil {
		return err
	}

	return core.Delete(ctx, core.DeleteOpts{
		RESTConfig: restConfig,
		Object:     el,
	})
}

func PatchAndDeleteMergeType(ctx context.Context, restConfig *rest.Config, el *unstructured.Unstructured) error {
	err := core.Patch(ctx, core.PatchOpts{
		RESTConfig: restConfig,
		GVK:        el.GroupVersionKind(),
		PatchData:  []byte(`{"metadata":{"finalizers":[]}}`),
		PatchType:  types.MergePatchType,
		Name:       el.GetName(),
		Namespace:  el.GetNamespace(),
	})
	if err != nil {
		return err
	}

	return core.Delete(ctx, core.DeleteOpts{
		RESTConfig: restConfig,
		Object:     el,
	})
}
