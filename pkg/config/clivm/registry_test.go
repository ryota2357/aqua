package clivm_test

import (
	"testing"

	"github.com/clivm/clivm/pkg/config/clivm"
)

func TestRegistry_Validate(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title    string
		registry *clivm.Registry
		isErr    bool
	}{
		{
			title: "github_content",
			registry: &clivm.Registry{
				RepoOwner: "clivm",
				RepoName:  "clivm-registry",
				Ref:       "v0.8.0",
				Path:      "foo.yaml",
				Type:      "github_content",
			},
		},
		{
			title: "github_content repo_owner is required",
			registry: &clivm.Registry{
				RepoName: "clivm-registry",
				Ref:      "v0.8.0",
				Path:     "foo.yaml",
				Type:     "github_content",
			},
			isErr: true,
		},
		{
			title: "github_content repo_name is required",
			registry: &clivm.Registry{
				RepoOwner: "clivm",
				Ref:       "v0.8.0",
				Path:      "foo.yaml",
				Type:      "github_content",
			},
			isErr: true,
		},
		{
			title: "github_content ref is required",
			registry: &clivm.Registry{
				RepoOwner: "clivm",
				RepoName:  "clivm-registry",
				Path:      "foo.yaml",
				Type:      "github_content",
			},
			isErr: true,
		},
		{
			title: "local",
			registry: &clivm.Registry{
				Path: "foo.yaml",
				Type: "local",
			},
		},
		{
			title: "local path is required",
			registry: &clivm.Registry{
				Type: "local",
			},
			isErr: true,
		},
		{
			title: "invalid type",
			registry: &clivm.Registry{
				Type: "invalid-type",
			},
			isErr: true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if err := d.registry.Validate(); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}

func TestLocalRegistry_GetFilePath(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		exp         string
		registry    *clivm.Registry
		rootDir     string
		cfgFilePath string
	}{
		{
			title:       "normal",
			exp:         "ci/foo.yaml",
			rootDir:     "/root/.aqua",
			cfgFilePath: "ci/clivm.yaml",
			registry: &clivm.Registry{
				Path: "foo.yaml",
				Type: "local",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if p := d.registry.GetFilePath(d.rootDir, d.cfgFilePath); p != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, p)
			}
		})
	}
}

func TestRegistry_GetFilePath(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		exp         string
		registry    *clivm.Registry
		rootDir     string
		cfgFilePath string
	}{
		{
			title:   "github_content",
			exp:     "/root/.aqua/registries/github_content/github.com/clivm/clivm-registry/v0.8.0/foo.yaml",
			rootDir: "/root/.aqua",
			registry: &clivm.Registry{
				RepoOwner: "clivm",
				RepoName:  "clivm-registry",
				Ref:       "v0.8.0",
				Path:      "foo.yaml",
				Type:      "github_content",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if p := d.registry.GetFilePath(d.rootDir, d.cfgFilePath); p != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, p)
			}
		})
	}
}