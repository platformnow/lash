package claims

import (
	"context"

	"github.com/platfornow/lash/internal/core"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (c Core) Apply(ctx context.Context, opts ModuleOpts) error {
	gvk := c.GetGroupVersionKind()

	obj := &unstructured.Unstructured{}
	obj.SetKind(gvk.Kind)
	obj.SetAPIVersion(gvk.GroupVersion().String())
	obj.SetName("core")
	obj.SetLabels(map[string]string{
		core.InstalledByLabel: core.InstalledByValue,
	})
	err := unstructured.SetNestedField(obj.Object, opts.Data, "spec")
	if err != nil {
		return err
	}

	return core.Apply(ctx, core.ApplyOpts{
		RESTConfig: opts.RESTConfig,
		GVK:        gvk,
		Object:     obj,
	})
}

func (c Core) GetGroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "pkg.platformnow.io",
		Version: "v1",
		Kind:    "Core",
	}
}
func (c Core) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "pkg.platformnow.io",
		Version:  "v1",
		Resource: "cores",
	}
}
