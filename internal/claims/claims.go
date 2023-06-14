package claims

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type ModuleOpts struct {
	RESTConfig *rest.Config
	Data       map[string]interface{}
}

type DeleteOpts struct {
	RESTConfig *rest.Config
	Name       string
}

type ManagedResource interface {
	Apply(ctx context.Context, opts ModuleOpts) error
	Delete(ctx context.Context, opts DeleteOpts) (err error)
	GetGroupVersionKind() schema.GroupVersionKind
	GetGroupVersionResource() schema.GroupVersionResource
	getName() string
}

type Core struct {
	name string
}

func NewCore(name string) *Core {
	return &Core{
		name: name,
	}
}

func (c Core) getName() string {
	return c.name
}

type GitOps struct {
	name string
}

func NewGitops(name string) *GitOps {
	return &GitOps{
		name: name,
	}
}

func (c GitOps) getName() string {
	return c.name
}

func (c GitOps) SetName(s string) {
	c.name = s
}

func ApplyModule(ctx context.Context, opts ModuleOpts, resource ManagedResource) error {
	return resource.Apply(ctx, opts)
}

func DeleteModule(ctx context.Context, opts DeleteOpts, resource ManagedResource) error {
	return resource.Delete(ctx, opts)
}
