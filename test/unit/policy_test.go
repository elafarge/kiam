// Copyright 2017 uSwitch
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
package kiam

import (
	"context"
	"testing"

	"github.com/uswitch/kiam/pkg/k8s"
	"github.com/uswitch/kiam/pkg/server"
	"github.com/uswitch/kiam/pkg/testutil"
)

func TestRequestedRolePolicy(t *testing.T) {
	p := testutil.NewPodWithRole("namespace", "name", "192.168.0.1", testutil.PhaseRunning, "myrole")
	f := testutil.NewStubFinder(p)

	policy := server.NewRequestingAnnotatedRolePolicy(f)
	decision, err := policy.IsAllowedAssumeRole(context.Background(), "myrole", "192.168.0.1")
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !decision.IsAllowed() {
		t.Error("role was same, should have been permitted:", decision.Explanation())
	}

	decision, _ = policy.IsAllowedAssumeRole(context.Background(), "wrongrole", "192.168.0.1")
	if decision.IsAllowed() {
		t.Error("role is different, should be denied", decision.Explanation())
	}
}

func TestErrorWhenPodNotFound(t *testing.T) {
	f := testutil.NewStubFinder(nil)
	policy := server.NewRequestingAnnotatedRolePolicy(f)

	_, err := policy.IsAllowedAssumeRole(context.Background(), "myrole", "192.168.0.1")
	if err == nil {
		t.Error("no pod found, should have been error")
	}

	if err != k8s.ErrPodNotFound {
		t.Error("wrong message", err.Error())
	}
}

func TestNamespacePolicy(t *testing.T) {
	n := testutil.NewNamespace("red", "^red.*$")
	nf := testutil.NewNamespaceFinder(n)
	p := testutil.NewPodWithRole("red", "foo", "192.168.0.1", testutil.PhaseRunning, "red_role")
	pf := testutil.NewStubFinder(p)

	policy := server.NewNamespacePermittedRoleNamePolicy(nf, pf)
	decision, err := policy.IsAllowedAssumeRole(context.Background(), "red_role", "192.168.0.1")
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !decision.IsAllowed() {
		t.Errorf("expected to be allowed- pod in correct namespace")
	}

	decision, _ = policy.IsAllowedAssumeRole(context.Background(), "orange_role", "192.168.0.1")
	if decision.IsAllowed() {
		t.Errorf("expected to be forbidden- requesting role that fails regexp")
	}
}

func TestNotAllowedWithoutNamespaceAnnotation(t *testing.T) {
	n := testutil.NewNamespace("red", "")
	nf := testutil.NewNamespaceFinder(n)
	p := testutil.NewPodWithRole("red", "foo", "192.168.0.1", testutil.PhaseRunning, "red_role")
	pf := testutil.NewStubFinder(p)

	policy := server.NewNamespacePermittedRoleNamePolicy(nf, pf)
	decision, _ := policy.IsAllowedAssumeRole(context.Background(), "red_role", "192.168.0.1")

	if decision.IsAllowed() {
		t.Error("expected failure, empty namespace policy annotation")
	}
}
