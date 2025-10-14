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

// Package utils contain extra functions used in _test.go files
package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
)

// Run will execute command and grab its output
func Run(cmd *exec.Cmd) (err error) {
	var dir string
	if dir, err = GetProjectDir(); err != nil {
		return err
	}
	cmd.Dir = dir

	if err = os.Chdir(cmd.Dir); err != nil {
		fmt.Fprintf(GinkgoWriter, "chdir dir: %q\n", err)
	}

	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	command := strings.Join(cmd.Args, " ")
	fmt.Fprintf(GinkgoWriter, "running: %q\n", command)
	var output []byte
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(GinkgoWriter, string(output))
		return fmt.Errorf("%q failed with error: %w", command, err)
	}
	return
}
