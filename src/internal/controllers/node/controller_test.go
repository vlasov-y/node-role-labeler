// Copyright 2025 The Node Role Labeler Authors
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

package node

import (
	"encoding/json"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/vlasov-y/node-role-labeler/internal/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("NodeReconciler", func() {
	var (
		node    *corev1.Node
		nodeKey types.NamespacedName
		req     reconcile.Request
	)

	BeforeEach(func() {
		node = &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
				Labels: map[string]string{
					"node-role.kubernetes.io/a": "",
					"node-role.cluster.local/b": "",
					"node-role.cluster.local/c": "value",
				},
			},
		}
		nodeKey = types.NamespacedName{Name: node.Name}
		req = reconcile.Request{NamespacedName: nodeKey}

		Expect(c.Create(ctx, node)).To(Succeed())
		Expect(c.Status().Update(ctx, node)).To(Succeed())
		Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.Unsetenv("NODE_ROLE_PREFIX")).To(Succeed())
		Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
		node.SetFinalizers([]string{})
		Expect(c.Update(ctx, node)).To(Succeed())
		if node.GetDeletionTimestamp() == nil {
			Expect(c.Delete(ctx, node)).To(Succeed())
		}
	})

	Context("when reconciling a node", func() {
		verifyState := func() {
			By("verifying the state")
			Expect(node.Annotations).To(HaveKey(AnnotationState))
			state := map[string]string{}
			Expect(json.Unmarshal([]byte(node.Annotations[AnnotationState]), &state)).To(Succeed())
			Expect(state).To(HaveKey("a"))
			Expect(state).To(HaveKey("b"))
			Expect(state).To(HaveKey("c"))
		}

		BeforeEach(func() {
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))
			Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
			Expect(recorder.Events).To(HaveLen(3))
			for range 3 {
				Expect(<-recorder.Events).To(ContainSubstring("RoleAdded"))
			}
		})

		AfterEach(func() {
			By("verifying the state")
			verifyState()
		})

		It("should duplicate official role to the custom one", func() {
			Expect(node.Labels).To(HaveKeyWithValue("node-role.cluster.local/a", ""))
		})

		It("should duplicate custom role to the official one", func() {
			Expect(node.Labels).To(HaveKeyWithValue("node-role.kubernetes.io/b", ""))
		})

		It("should preserve duplicated label value", func() {
			Expect(node.Labels).To(HaveKeyWithValue("node-role.kubernetes.io/c", "value"))
		})

		It("should handle deleted state", func() {
			By("deleting the state")
			delete(node.Labels, AnnotationState)
			Expect(c.Update(ctx, node)).To(Succeed())

			By("reconciling")
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))
			Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
			Expect(recorder.Events).To(BeEmpty())
		})

		It("should handle a bad state", func() {
			By("preparing a bad state")
			state := map[string]string{
				"a":                   "bad",
				"b":                   "bad",
				"c":                   "bad",
				"should not be there": "but it is",
			}
			stateMarshaled, err := json.Marshal(state)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())

			By("updating the state")
			node.Annotations[AnnotationState] = string(stateMarshaled)
			Expect(c.Update(ctx, node)).To(Succeed())

			By("reconciling")
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))
			Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
			Expect(recorder.Events).To(HaveLen(3))
			for range 3 {
				Expect(<-recorder.Events).To(ContainSubstring("StateRepair"))
			}
		})

		It("should handle a role value change", func() {
			By("updating a role value")
			node.Labels["node-role.kubernetes.io/a"] = "new"
			node.Labels["node-role.cluster.local/b"] = "new"
			Expect(c.Update(ctx, node)).To(Succeed())

			By("reconciling")
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))
			Expect(c.Get(ctx, nodeKey, node)).To(Succeed())

			By("verifying new role values")
			Expect(node.Labels).To(HaveKeyWithValue("node-role.cluster.local/a", "new"))
			Expect(node.Labels).To(HaveKeyWithValue("node-role.kubernetes.io/b", "new"))

			Expect(recorder.Events).To(HaveLen(2))
			for range 2 {
				Expect(<-recorder.Events).To(ContainSubstring("RoleChanged"))
			}
		})
	})

	Context("when reconciling an excluded node", func() {
		BeforeEach(func() {
			node.SetAnnotations(map[string]string{
				AnnotationEnable: "false",
			})
			Expect(c.Update(ctx, node)).To(Succeed())
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))
			Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
		})

		It("should not alter any labels", func() {
			Expect(node.Annotations).To(HaveLen(1))
			Expect(node.Labels).To(HaveLen(3))
		})
	})

	Context("when node role prefix is overridden", func() {
		BeforeEach(func() {
			Expect(os.Setenv("NODE_ROLE_PREFIX", "overridden/")).To(Succeed())
			labels := map[string]string{}
			for k, v := range node.Labels {
				labels[strings.ReplaceAll(k, "node-role.cluster.local", "overridden")] = v
			}
			node.SetLabels(labels)
			Expect(c.Update(ctx, node)).To(Succeed())
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))
			Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
			Expect(recorder.Events).To(HaveLen(3))
			for range 3 {
				Expect(<-recorder.Events).To(ContainSubstring("RoleAdded"))
			}
		})

		It("should duplicate official role to the custom one", func() {
			Expect(node.Labels).To(HaveKeyWithValue("overridden/a", ""))
		})

		It("should duplicate custom role to the official one", func() {
			Expect(node.Labels).To(HaveKeyWithValue("node-role.kubernetes.io/b", ""))
		})

		It("should preserve duplicated label value", func() {
			Expect(node.Labels).To(HaveKeyWithValue("node-role.kubernetes.io/c", "value"))
		})
	})

	Context("when reserved prefix is used", func() {
		It("should reconcile without an error but with log", func() {
			Expect(os.Setenv("NODE_ROLE_PREFIX", "node-role.kubernetes.io/")).To(Succeed())
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).To(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))
			Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
		})
	})

	Context("when node is being deleted", func() {
		BeforeEach(func() {
			node.SetFinalizers(append(node.GetFinalizers(), "unit.test/finalizer"))
			Expect(c.Update(ctx, node)).To(Succeed())
			Expect(c.Delete(ctx, node)).To(Succeed())
		})

		It("should handle deletion gracefully", func() {
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))
		})
	})

})
