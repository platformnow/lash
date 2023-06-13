package claims

import (
	"context"

	"github.com/platfornow/lash/internal/core"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (g GitOps) Apply(ctx context.Context, opts ModuleOpts) error {
	gvk := g.GetGroupVersionKind()

	obj := &unstructured.Unstructured{}
	obj.SetKind(gvk.Kind)
	obj.SetAPIVersion(gvk.GroupVersion().String())
	obj.SetName("gitops")
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

func (g GitOps) GetGroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "pkg.platformnow.io",
		Version: "v1",
		Kind:    "Gitops",
	}
}

func (g GitOps) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "pkg.platformnow.io",
		Version:  "v1",
		Resource: "gitops",
	}
}
