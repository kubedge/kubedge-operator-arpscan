// Copyright 2019 The Kubedge Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package arpscan

import (
	"context"
	//JEB "reflect"

	av1 "github.com/kubedge/kubedge-operator-base/pkg/apis/kubedgeoperators/v1alpha1"
	bmgr "github.com/kubedge/kubedge-operator-base/pkg/kubedgemanager"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type rollbackmanager struct {
	bmgr.KubedgeBaseManager

	spec   av1.ArpscanSpec
	status *av1.ArpscanStatus
}

// Sync retrieves from K8s the sub resources (Workflow, Job, ....) attached to this Arpscan CR
func (m *rollbackmanager) Sync(ctx context.Context) error {
	return m.BaseSync(ctx)
}

// InstallResource creates K8s sub resources (Workflow, Job, ....) attached to this Arpscan CR
func (m rollbackmanager) InstallResource(ctx context.Context) (*av1.SubResourceList, error) {
	return m.BaseInstallResource(ctx)
}

// InstallResource updates K8s sub resources (Workflow, Job, ....) attached to this Arpscan CR
func (m rollbackmanager) UpdateResource(ctx context.Context) (*av1.SubResourceList, *av1.SubResourceList, error) {
	return m.BaseUpdateResource(ctx)
}

// ReconcileResource creates or patches resources as necessary to match this Arpscan CR
func (m rollbackmanager) ReconcileResource(ctx context.Context) (*av1.SubResourceList, error) {
	return m.BaseReconcileResource(ctx)
}

// UninstallResource delete K8s sub resources (Workflow, Job, ....) attached to this Arpscan CR
func (m rollbackmanager) UninstallResource(ctx context.Context) (*av1.SubResourceList, error) {
	return m.BaseUninstallResource(ctx)
}

// deploymentForKubedge returns a kubedge Deployment object
func (m rollbackmanager) deploymentForKubedge(instance *av1.Arpscan) *appsv1.Deployment {
	// size := instance.Spec.Size
	// if *found.Spec.Replicas != size

	ls := m.labelsForKubedge(instance.Name)
	// replicas := instance.Spec.Size
	replicas := int32(1)

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   "hack4easy/arpscan-amd64:latest",
						Name:    "arpscan",
						Command: []string{"/bin/arpscan", "eth0"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "arpscan",
						}},
					}},
				},
			},
		},
	}
	return dep
}

// labelsForKubedge returns the labels for selecting the resources
// belonging to the given kubedge CR name.
func (m rollbackmanager) labelsForKubedge(name string) map[string]string {
	return map[string]string{"app": "kubedge", "kubedge_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func (m rollbackmanager) getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
