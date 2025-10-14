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

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/vlasov-y/node-role-labeler/internal/types"
	. "github.com/vlasov-y/node-role-labeler/test/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/describe"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespace = "node-role-labeler"
)

var _ = Describe("Manager", Ordered, func() {
	var podName string
	var c client.Client
	var cs *kubernetes.Clientset
	var rc *rest.Config
	var err error
	var cmd *exec.Cmd
	var ctx context.Context
	var cancel context.CancelFunc

	// Before running the tests, set up the environment by creating the namespace,
	// enforce the restricted security policy to the namespace, installing CRDs,
	// and deploying the controller.
	BeforeAll(func() {
		ctx, cancel = context.WithCancel(context.Background())

		By("installing operator to the cluster")
		cmd = exec.Command("task", "install-operator")
		err = Run(cmd)
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to install the operator")

		By("initializing Kubernetes clients")
		var cwd string
		cwd, err = GetProjectDir()
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to get project dir")
		rc, err = clientcmd.BuildConfigFromFlags("", path.Join(cwd, "kubeconfig.yaml"))
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to load kubeconfig")
		c, err = client.New(rc, client.Options{})
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to create the client")
		cs, err = kubernetes.NewForConfig(rc)
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to create the clientset")

		By("finding operator pod name")
		var pod corev1.Pod
		pods := &corev1.PodList{}
		err = c.List(context.Background(), pods, client.InNamespace(namespace))
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to list pods in the namespace")
		Expect(pods.Items).To(HaveLen(1), "expected eactly 1 pod in the namespace")
		pod = pods.Items[0]
		podName = pod.Name
	})

	// After all tests have been executed, clean up by undeploying the controller, uninstalling CRDs,
	// and deleting the namespace.
	AfterAll(func() {
		By("uninstalling the operator")
		cmd = exec.Command("task", "uninstall-operator")
		err = Run(cmd)
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to uninstall the operator")
		cancel()
	})

	// After each test, check for failures and collect logs, events,
	// and pod descriptions for debugging.
	AfterEach(func() {
		specReport := CurrentSpecReport()
		if specReport.Failed() {
			By("Fetching controller manager pod logs")
			logs, err := cs.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{}).DoRaw(context.Background())
			ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to get Controller logs")
			fmt.Fprintf(GinkgoWriter, ">>>\n>>> Controller logs\n>>>\n %s\n", string(logs))

			By("Fetching Kubernetes events")
			events := &corev1.EventList{}
			err = c.List(context.Background(), events, client.InNamespace(namespace))
			ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to get Kubernetes events")
			fmt.Fprintln(GinkgoWriter, ">>>\n>>> Kubernetes Events\n>>>")
			for _, event := range events.Items {
				fmt.Fprintf(GinkgoWriter, "%s | %s | %s\n", event.LastTimestamp.Time, event.Reason, event.Message)
			}

			By("Describing operator pod")
			fmt.Fprintln(GinkgoWriter, ">>>\n>>> Pod Describe\n>>>")
			describer := describe.PodDescriber{Interface: cs}
			output, err := describer.Describe(namespace, podName, describe.DescriberSettings{
				ShowEvents: true,
			})
			ExpectWithOffset(2, err).NotTo(HaveOccurred(), "failed to describe the pod")
			fmt.Fprintln(GinkgoWriter, output)
		}
	})

	Context("managing node roles", func() {
		var node *corev1.Node
		var nodeKey types.NamespacedName

		SetDefaultEventuallyPollingInterval(time.Millisecond * 200)
		SetDefaultEventuallyTimeout(time.Second * 3)

		BeforeEach(func() {
			nodes := corev1.NodeList{}
			Expect(c.List(ctx, &nodes)).To(Succeed())
			Expect(nodes.Items).ToNot(BeEmpty())
			node = &nodes.Items[0]
			Expect(node.Labels).ToNot(BeEmpty())
			nodeKey = types.NamespacedName{Name: node.Name}

		})

		It("should handle adding roles", func() {
			node.Labels["node-role.cluster.local/e2e"] = "e2e"
			Expect(c.Update(ctx, node)).To(Succeed())
			Eventually(func() bool {
				Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
				if value, exists := node.Labels["node-role.kubernetes.io/e2e"]; exists && value == "e2e" {
					return true
				}
				return false
			}).Should(BeTrue())
		})

		It("should handle altering roles", func() {
			node.Labels["node-role.cluster.local/e2e"] = "new"
			Expect(c.Update(ctx, node)).To(Succeed())
			Eventually(func() bool {
				Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
				if value, exists := node.Labels["node-role.kubernetes.io/e2e"]; exists && value == "new" {
					return true
				}
				return false
			}).Should(BeTrue())
		})

		It("should handle deleting state", func() {
			Expect(node.Annotations).ToNot(BeEmpty())
			delete(node.Annotations, AnnotationState)
			Expect(c.Update(ctx, node)).To(Succeed())
			Eventually(func() bool {
				Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
				stateStr, stateExists := node.Annotations[AnnotationState]
				if !stateExists {
					return false
				}
				state := map[string]string{}
				if err := json.Unmarshal([]byte(stateStr), &state); err != nil {
					return false
				}
				if value, exists := state["e2e"]; !exists || value == "" {
					return false
				}
				return true
			}).Should(BeTrue())
		})

		It("should handle deleting roles", func() {
			delete(node.Labels, "node-role.cluster.local/e2e")
			Expect(c.Update(ctx, node)).To(Succeed())
			Eventually(func() bool {
				Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
				_, customExists := node.Labels["node-role.cluster.local/e2e"]
				_, officialExists := node.Labels["node-role.kubernetes.io/e2e"]
				if !customExists && !officialExists {
					return true
				}
				return false
			}).Should(BeTrue())
		})

	})
})
