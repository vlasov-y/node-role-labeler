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
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/vlasov-y/node-role-labeler/internal/types"
	"github.com/vlasov-y/node-role-labeler/test/utils"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	c          client.Client
	ctx        context.Context
	reconciler *NodeReconciler
	recorder   *record.FakeRecorder
	suite      *utils.ControllerTestSuite
)

func TestNode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node Controller")
}

var _ = BeforeSuite(func() {
	suite = utils.NewControllerTestSuite()
	ExpectWithOffset(1, suite).ToNot(BeNil())
	// Just easier to reach in tests, less text
	c = suite.Client
	ctx = suite.Ctx
	recorder = suite.Recorder

	reconciler = &NodeReconciler{
		Reconciler: Reconciler{
			Client:   suite.Client,
			Config:   suite.Config,
			Scheme:   suite.Client.Scheme(),
			Recorder: suite.Recorder,
		},
	}
})

var _ = AfterSuite(func() {
	suite.Teardown()
})
