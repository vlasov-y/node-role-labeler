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
