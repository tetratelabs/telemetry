// Copyright 2022 Tetrate
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

package group_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/tetratelabs/log"
	"github.com/tetratelabs/run"

	"github.com/tetratelabs/telemetry"
	"github.com/tetratelabs/telemetry/group"
	"github.com/tetratelabs/telemetry/scope"
)

func TestService(t *testing.T) {
	tests := []struct {
		name          string
		expectedLines []string
		run           func(l telemetry.Logger)
	}{
		{
			// We use test.name to initialize level.
			"info",
			[]string{
				" info 	test v0.0.0-unofficial started",
				" info 	ok",
				" info 	haha",
			},
			func(l telemetry.Logger) {
				l.Info("ok")
				l.Info("haha")
			},
		},
		{
			"debug",
			[]string{
				" info 	test v0.0.0-unofficial started",
				" debug	ok",
				" debug	haha",
			},
			func(l telemetry.Logger) {
				l.Debug("ok")
				l.Debug("haha")
			},
		},
	}

	scopeName := func(i int) string {
		return fmt.Sprintf("test%d", i)
	}

	// Register all possible scopes. Since UseLogger will register all possible scopes and can't be
	// changed.
	for i, test := range tests {
		scope.Register(scopeName(i), test.name)
	}

	tmp, err := ioutil.TempFile("", "log_test")
	if err != nil {
		t.Fatal(err)
	}
	oldStdout := os.Stdout
	// Redirect stdout to tmp.
	os.Stdout = tmp
	defer func() {
		_ = os.Remove(tmp.Name())
		os.Stdout = oldStdout
	}()

	defaultLogger := log.NewUnstructured()
	for i, test := range tests {
		var (
			g   = &run.Group{Name: "test", Logger: defaultLogger}
			svc = group.New(defaultLogger)
		)
		g.Register(svc)

		oldArgs := os.Args
		// Set current scope output level.
		os.Args = []string{"cmd", "--log-output-level=" + scopeName(i) + ":" + test.name}
		defer func() {
			os.Args = oldArgs
		}()

		if err := g.RunConfig(); err != nil {
			t.Fatalf("configuring run.Group: %v", err)
		}
		test.run(defaultLogger)

		content, _ := os.ReadFile(tmp.Name())
		_ = os.WriteFile(tmp.Name(), []byte{}, os.ModePerm)

		lines := strings.Split(string(content), "\n")

		for i, expectedLine := range test.expectedLines {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				entries := strings.SplitN(lines[i], " ", 3)
				entry := entries[len(entries)-1]
				if entry != expectedLine {
					t.Errorf("got '%s', expecting to equal '%s'", entry, expectedLine)
				}
			})
		}
	}
}
