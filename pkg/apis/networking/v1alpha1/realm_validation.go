/*
Copyright 2019 The Knative Authors

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

package v1alpha1

import (
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	"knative.dev/pkg/apis"
)

// Validate inspects and validates Ingress object.
func (r *Realm) Validate(ctx context.Context) *apis.FieldError {
	return r.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec")
}

// Validate inspects and validates IngressSpec object.
func (spec *RealmSpec) Validate(ctx context.Context) *apis.FieldError {
	// Spec must not be empty.
	if equality.Semantic.DeepEqual(spec, &RealmSpec{}) {
		return apis.ErrMissingField(apis.CurrentField)
	}
	var all *apis.FieldError
	if spec.External == "" && spec.Cluster == "" {
		all = all.Also(apis.ErrMissingOneOf("External", "Cluster"))
	}
	return all
}
