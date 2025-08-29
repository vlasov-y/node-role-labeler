/*
Copyright 2025.

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

package controller

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("NodeReconciler", func() {
	var (
		reconciler *NodeReconciler
		ctx        context.Context
		scheme     *runtime.Scheme
		client     *fake.ClientBuilder
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		client = fake.NewClientBuilder().WithScheme(scheme)
		reconciler = &NodeReconciler{
			Scheme:   scheme,
			Recorder: record.NewFakeRecorder(10),
		}
	})

	AfterEach(func() {
		os.Unsetenv("NODE_ROLE_PREFIX")
	})

	Context("when duplicating role labels", func() {
		It("should duplicate custom role to official", func() {
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"node-role.cluster.local/worker": "",
					},
				},
			}

			reconciler.Client = client.WithObjects(node).Build()
			req := ctrl.Request{NamespacedName: types.NamespacedName{Name: "test-node"}}
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			var updatedNode corev1.Node
			Expect(reconciler.Client.Get(ctx, types.NamespacedName{Name: "test-node"}, &updatedNode)).To(Succeed())
			Expect(updatedNode.Labels).To(HaveKeyWithValue("node-role.cluster.local/worker", ""))
			Expect(updatedNode.Labels).To(HaveKeyWithValue("node-role.kubernetes.io/worker", ""))
		})

		It("should duplicate official role to custom", func() {
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"node-role.kubernetes.io/master": "",
					},
				},
			}

			reconciler.Client = client.WithObjects(node).Build()
			req := ctrl.Request{NamespacedName: types.NamespacedName{Name: "test-node"}}
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			var updatedNode corev1.Node
			Expect(reconciler.Client.Get(ctx, types.NamespacedName{Name: "test-node"}, &updatedNode)).To(Succeed())
			Expect(updatedNode.Labels).To(HaveKeyWithValue("node-role.cluster.local/master", ""))
			Expect(updatedNode.Labels).To(HaveKeyWithValue("node-role.kubernetes.io/master", ""))
		})
	})

	Context("when enable annotation is false", func() {
		It("should skip processing", func() {
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"node-role.cluster.local/worker": "",
					},
					Annotations: map[string]string{
						"node-role-labeler.io/enable": "false",
					},
				},
			}

			reconciler.Client = client.WithObjects(node).Build()
			req := ctrl.Request{NamespacedName: types.NamespacedName{Name: "test-node"}}
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			var updatedNode corev1.Node
			Expect(reconciler.Client.Get(ctx, types.NamespacedName{Name: "test-node"}, &updatedNode)).To(Succeed())
			Expect(updatedNode.Labels).To(HaveKeyWithValue("node-role.cluster.local/worker", ""))
			Expect(updatedNode.Labels).NotTo(HaveKey("node-role.kubernetes.io/worker"))
		})
	})

	Context("when using custom prefix from env", func() {
		BeforeEach(func() {
			os.Setenv("NODE_ROLE_PREFIX", "node-role.example.com/")
		})

		It("should use custom prefix", func() {
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"node-role.example.com/worker": "",
					},
				},
			}

			reconciler.Client = client.WithObjects(node).Build()
			req := ctrl.Request{NamespacedName: types.NamespacedName{Name: "test-node"}}
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			var updatedNode corev1.Node
			Expect(reconciler.Client.Get(ctx, types.NamespacedName{Name: "test-node"}, &updatedNode)).To(Succeed())
			Expect(updatedNode.Labels).To(HaveKeyWithValue("node-role.example.com/worker", ""))
			Expect(updatedNode.Labels).To(HaveKeyWithValue("node-role.kubernetes.io/worker", ""))
		})
	})

	Context("when reserved prefix is used", func() {
		BeforeEach(func() {
			os.Setenv("NODE_ROLE_PREFIX", "node-role.kubernetes.io/")
		})

		It("should return error", func() {
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
			}

			reconciler.Client = client.WithObjects(node).Build()
			req := ctrl.Request{NamespacedName: types.NamespacedName{Name: "test-node"}}
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).To(HaveOccurred())
		})
	})
})
