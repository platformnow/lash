package managed

import (
	"context"

	"github.com/platfornow/lash/internal/core"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func ListK8Objects(ctx context.Context, restConfig *rest.Config) ([]unstructured.Unstructured, error) {
	return core.List(ctx, core.ListOpts{
		RESTConfig: restConfig,
		GVK: schema.GroupVersionKind{
			Group:   "kubernetes.crossplane.io",
			Version: "v1alpha1",
			Kind:    "Object",
		},
	})
}
