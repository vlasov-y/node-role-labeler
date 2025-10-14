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

package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("NewControllerTestSuite", Ordered, func() {
	var suite *ControllerTestSuite

	BeforeAll(func() {
		suite = NewControllerTestSuite()
	})

	AfterAll(func() {
		suite.Teardown()
	})

	Context("when setting up test suite", func() {
		It("should have all object initialized", func() {
			Expect(suite.Client).ToNot(BeNil())
			Expect(suite.Config).ToNot(BeNil())
			Expect(suite.Ctx).ToNot(BeNil())
			Expect(suite.Recorder).ToNot(BeNil())
			Expect(suite.TestEnv).ToNot(BeNil())
		})

		It("should allow creating and retrieving nodes", func() {
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			}
			Expect(suite.Client.Create(suite.Ctx, node)).To(Succeed())
			retrieved := &corev1.Node{}
			key := types.NamespacedName{Name: "test"}
			Expect(suite.Client.Get(suite.Ctx, key, retrieved)).To(Succeed())
			Expect(retrieved.Name).To(Equal("test"))
			Expect(suite.Client.Delete(suite.Ctx, node)).To(Succeed())
		})

		It("should allow listing cluster resources", func() {
			nodeList := &corev1.NodeList{}
			Expect(suite.Client.List(suite.Ctx, nodeList)).To(Succeed())
		})
	})
})
