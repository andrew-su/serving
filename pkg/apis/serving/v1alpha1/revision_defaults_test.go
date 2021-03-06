/*
Copyright 2018 The Knative Authors

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
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logtesting "knative.dev/pkg/logging/testing"
	"knative.dev/pkg/ptr"

	"knative.dev/serving/pkg/apis/config"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
	autoscalerconfig "knative.dev/serving/pkg/autoscaler/config"
)

var defaultProbe = &corev1.Probe{
	SuccessThreshold: 1,
	Handler: corev1.Handler{
		TCPSocket: &corev1.TCPSocketAction{},
	},
}

func TestRevisionDefaulting(t *testing.T) {
	tests := []struct {
		name string
		in   *Revision
		want *Revision
		wc   func(context.Context) context.Context
	}{{
		name: "empty",
		in:   &Revision{},
		want: &Revision{
			Spec: RevisionSpec{
				RevisionSpec: v1.RevisionSpec{
					ContainerConcurrency: ptr.Int64(0),
					TimeoutSeconds:       ptr.Int64(config.DefaultRevisionTimeoutSeconds),
				},
				DeprecatedContainer: &corev1.Container{
					Name:           config.DefaultUserContainerName,
					Resources:      defaultResources,
					ReadinessProbe: defaultProbe,
				},
			},
		},
	}, {
		name: "shell",
		in: &Revision{
			Spec: RevisionSpec{
				DeprecatedContainer: &corev1.Container{},
			},
		},
		want: &Revision{
			Spec: RevisionSpec{
				RevisionSpec: v1.RevisionSpec{
					ContainerConcurrency: ptr.Int64(0),
					TimeoutSeconds:       ptr.Int64(config.DefaultRevisionTimeoutSeconds),
				},
				DeprecatedContainer: &corev1.Container{
					Name:           config.DefaultUserContainerName,
					Resources:      defaultResources,
					ReadinessProbe: defaultProbe,
				},
			},
		},
	}, {
		name: "with context",
		in: &Revision{
			Spec: RevisionSpec{
				DeprecatedContainer: &corev1.Container{},
			}},
		wc: func(ctx context.Context) context.Context {
			s := config.NewStore(logtesting.TestLogger(t))
			s.OnConfigChanged(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: autoscalerconfig.ConfigName}})
			s.OnConfigChanged(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: config.FeaturesConfigName}})
			s.OnConfigChanged(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: config.DefaultsConfigName,
				},
				Data: map[string]string{
					"revision-timeout-seconds": "123",
				},
			})

			return s.ToContext(ctx)
		},
		want: &Revision{
			Spec: RevisionSpec{
				RevisionSpec: v1.RevisionSpec{
					ContainerConcurrency: ptr.Int64(0),
					TimeoutSeconds:       ptr.Int64(123),
				},
				DeprecatedContainer: &corev1.Container{
					Name:           config.DefaultUserContainerName,
					Resources:      defaultResources,
					ReadinessProbe: defaultProbe,
				},
			},
		},
	}, {
		name: "readonly volumes",
		in: &Revision{
			Spec: RevisionSpec{
				DeprecatedContainer: &corev1.Container{
					Image: "foo",
					VolumeMounts: []corev1.VolumeMount{{
						Name: "bar",
					}},
				},
				RevisionSpec: v1.RevisionSpec{
					ContainerConcurrency: ptr.Int64(1),
					TimeoutSeconds:       ptr.Int64(99),
				},
			},
		},
		want: &Revision{
			Spec: RevisionSpec{
				DeprecatedContainer: &corev1.Container{
					Name:  config.DefaultUserContainerName,
					Image: "foo",
					VolumeMounts: []corev1.VolumeMount{{
						Name:     "bar",
						ReadOnly: true,
					}},
					Resources:      defaultResources,
					ReadinessProbe: defaultProbe,
				},
				RevisionSpec: v1.RevisionSpec{
					ContainerConcurrency: ptr.Int64(1),
					TimeoutSeconds:       ptr.Int64(99),
				},
			},
		},
	}, {
		name: "lemonade",
		wc:   v1.WithUpgradeViaDefaulting,
		in: &Revision{
			Spec: RevisionSpec{
				DeprecatedContainer: &corev1.Container{
					Image: "foo",
					VolumeMounts: []corev1.VolumeMount{{
						Name: "bar",
					}},
				},
				RevisionSpec: v1.RevisionSpec{
					ContainerConcurrency: ptr.Int64(1),
					TimeoutSeconds:       ptr.Int64(99),
				},
			},
		},
		want: &Revision{
			Spec: RevisionSpec{
				RevisionSpec: v1.RevisionSpec{
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  config.DefaultUserContainerName,
							Image: "foo",
							VolumeMounts: []corev1.VolumeMount{{
								Name:     "bar",
								ReadOnly: true,
							}},
							Resources:      defaultResources,
							ReadinessProbe: defaultProbe,
						}},
					},
					ContainerConcurrency: ptr.Int64(1),
					TimeoutSeconds:       ptr.Int64(99),
				},
			},
		},
	}, {
		name: "lemonade (no overwrite)",
		wc:   v1.WithUpgradeViaDefaulting,
		in: &Revision{
			Spec: RevisionSpec{
				DeprecatedContainer: &corev1.Container{
					Image: "bar",
				},
				RevisionSpec: v1.RevisionSpec{
					ContainerConcurrency: ptr.Int64(1),
					TimeoutSeconds:       ptr.Int64(99),
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image:          "foo",
							Resources:      defaultResources,
							ReadinessProbe: defaultProbe,
						}},
					},
				},
			},
		},
		want: &Revision{
			Spec: RevisionSpec{
				DeprecatedContainer: &corev1.Container{
					Image: "bar",
				},
				RevisionSpec: v1.RevisionSpec{
					ContainerConcurrency: ptr.Int64(1),
					TimeoutSeconds:       ptr.Int64(99),
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:           config.DefaultUserContainerName,
							Image:          "foo",
							Resources:      defaultResources,
							ReadinessProbe: defaultProbe,
						}},
					},
				},
			},
		},
	}, {
		name: "no overwrite",
		in: &Revision{
			Spec: RevisionSpec{
				DeprecatedContainer: &corev1.Container{},
				RevisionSpec: v1.RevisionSpec{
					ContainerConcurrency: ptr.Int64(1),
					TimeoutSeconds:       ptr.Int64(99),
				},
			},
		},
		want: &Revision{
			Spec: RevisionSpec{
				RevisionSpec: v1.RevisionSpec{
					ContainerConcurrency: ptr.Int64(1),
					TimeoutSeconds:       ptr.Int64(99),
				},
				DeprecatedContainer: &corev1.Container{
					Name:           config.DefaultUserContainerName,
					Resources:      defaultResources,
					ReadinessProbe: defaultProbe,
				},
			},
		},
	}, {
		name: "partially initialized",
		in: &Revision{
			Spec: RevisionSpec{
				DeprecatedContainer: &corev1.Container{},
				RevisionSpec: v1.RevisionSpec{
					ContainerConcurrency: ptr.Int64(123),
				},
			},
		},
		want: &Revision{
			Spec: RevisionSpec{
				RevisionSpec: v1.RevisionSpec{
					ContainerConcurrency: ptr.Int64(123),
					TimeoutSeconds:       ptr.Int64(config.DefaultRevisionTimeoutSeconds),
				},
				DeprecatedContainer: &corev1.Container{
					Name:           config.DefaultUserContainerName,
					Resources:      defaultResources,
					ReadinessProbe: defaultProbe,
				},
			},
		},
	}, {
		name: "multiple containers",
		wc:   v1.WithUpgradeViaDefaulting,
		in: &Revision{
			Spec: RevisionSpec{
				RevisionSpec: v1.RevisionSpec{
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name: "busybox",
							Ports: []corev1.ContainerPort{{
								ContainerPort: 8888,
							}},
						}, {
							Name: "helloworld",
						}},
					},
					ContainerConcurrency: ptr.Int64(1),
					TimeoutSeconds:       ptr.Int64(99),
				},
			},
		},
		want: &Revision{
			Spec: RevisionSpec{
				RevisionSpec: v1.RevisionSpec{
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name: "busybox",
							Ports: []corev1.ContainerPort{{
								ContainerPort: 8888,
							}},
							Resources:      defaultResources,
							ReadinessProbe: defaultProbe,
						}, {
							Name:      "helloworld",
							Resources: defaultResources,
						}},
					},
					ContainerConcurrency: ptr.Int64(1),
					TimeoutSeconds:       ptr.Int64(99),
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.in
			ctx := context.Background()
			if test.wc != nil {
				ctx = test.wc(ctx)
			}
			got.SetDefaults(ctx)
			if diff := cmp.Diff(test.want, got, ignoreUnexportedResources); diff != "" {
				t.Errorf("SetDefaults (-want, +got) = %v", diff)
			}
		})
	}
}

func TestRevisionDefaultingContainerName(t *testing.T) {
	got := &Revision{
		Spec: RevisionSpec{
			RevisionSpec: v1.RevisionSpec{
				PodSpec: corev1.PodSpec{
					Containers: []corev1.Container{{}, {}},
				},
				ContainerConcurrency: ptr.Int64(1),
				TimeoutSeconds:       ptr.Int64(99),
			},
		},
	}
	got.SetDefaults(context.Background())
	if got.Spec.Containers[0].Name == "" && got.Spec.Containers[1].Name == "" {
		t.Errorf("Failed to set default values for container name")
	}
}
