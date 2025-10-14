package e2e

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/vlasov-y/node-role-labeler/test/utils"
)

// TestE2E runs the end-to-end (e2e) test suite for the project. These tests execute in an isolated,
// temporary environment to validate project changes with the purposed to be used in CI jobs.
// The default setup requires Kind, builds/loads the Manager Docker image locally, and installs
// CertManager.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "End-to-end")
}

var _ = BeforeSuite(func() {
	By("creating and bootstrapping a kind cluster")
	cmd := exec.Command("task", "kind:bootstrap")
	err := Run(cmd)
	ExpectWithOffset(1, err).ToNot(HaveOccurred(), "Failed to create the kind cluster")
})
