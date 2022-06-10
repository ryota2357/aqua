package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func TestPackageInfos_ToMap(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		pkgInfos *registry.PackageInfos
		exp      map[string]*registry.PackageInfo
		isErr    bool
	}{
		{
			title: "normal",
			pkgInfos: &registry.PackageInfos{
				&registry.PackageInfo{
					Type: "github_release",
					Name: "foo",
				},
			},
			exp: map[string]*registry.PackageInfo{
				"foo": {
					Type: "github_release",
					Name: "foo",
				},
			},
		},
	}

	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			m, err := d.pkgInfos.ToMap(logE)
			if d.isErr {
				if err == nil {
					t.Fatal("error should be returned")
				}
				return
			}
			if diff := cmp.Diff(d.exp, m); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
