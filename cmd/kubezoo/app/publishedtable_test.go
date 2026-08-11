/*
Copyright 2024 The KubeZoo Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"context"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/storage"
)

// TestPublishedStorageClassShowsTheDecisionColumns guards the columns a tenant
// reads to choose a class.
//
// ⛔ Without a real table convertor, kubectl get sc falls back to NAME and
// CREATED AT. That drops exactly the fields a tenant decides on: whether the
// volume is deleted or retained, and whether it can be expanded. The data is in
// the object either way (-o yaml has it), but a tenant reads the table, not the
// YAML, and a table that hides retain-vs-delete is a table that lets a tenant
// pick a class whose behaviour it could not see.
func TestPublishedStorageClassShowsTheDecisionColumns(t *testing.T) {
	sc := &storage.StorageClass{
		ObjectMeta:           metav1.ObjectMeta{Name: "ceph"},
		Provisioner:          "cinder.knaas.io",
		ReclaimPolicy:        reclaim(core.PersistentVolumeReclaimDelete),
		VolumeBindingMode:    bindingMode(storage.VolumeBindingImmediate),
		AllowVolumeExpansion: func() *bool { b := true; return &b }(),
	}
	table, err := publishedClassTableConvertor.ConvertToTable(context.TODO(), sc, nil)
	if err != nil {
		t.Fatalf("converting a StorageClass to a table: %v", err)
	}
	var headers []string
	for _, c := range table.ColumnDefinitions {
		headers = append(headers, c.Name)
	}
	joined := strings.Join(headers, ",")
	for _, want := range []string{"Provisioner", "ReclaimPolicy", "VolumeBindingMode", "AllowVolumeExpansion"} {
		if !strings.Contains(joined, want) {
			t.Errorf("column %q missing; a tenant cannot see it in `kubectl get sc`.\n"+
				"columns = %s\n\nthis is the default-table fallback (NAME + CREATED AT), "+
				"which drops the fields a tenant chooses a class on", want, joined)
		}
	}
}

func reclaim(p core.PersistentVolumeReclaimPolicy) *core.PersistentVolumeReclaimPolicy {
	return &p
}
func bindingMode(m storage.VolumeBindingMode) *storage.VolumeBindingMode { return &m }
