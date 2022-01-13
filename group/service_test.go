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
	scope.Register("ok", "ok")
	tests := []struct {
		name          string
		expectedLines []string
		run           func(l telemetry.Logger)
	}{
		{
			"test",
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
	}

	for _, test := range tests {
		lines, _ := captureStdout(func() {
			var (
				l   = log.NewUnstructured()
				g   = &run.Group{Name: test.name, Logger: l}
				svc = group.New(l)
			)
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{"cmd", "--log-output-level=ok:info"}
			g.Register(svc)

			if err := g.RunConfig(); err != nil {
				t.Fatalf("configuring run.Group: %v", err)
			}

			test.run(l)
		})

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

// captureStdout runs the given function while capturing everything sent to stdout.
func captureStdout(f func()) ([]string, error) {
	tf, err := ioutil.TempFile("", "log_test")
	if err != nil {
		return nil, err
	}

	old := os.Stdout
	os.Stdout = tf

	f()

	os.Stdout = old
	path := tf.Name()
	_ = tf.Sync()
	_ = tf.Close()

	content, err := os.ReadFile(path)
	_ = os.Remove(path)

	if err != nil {
		return nil, err
	}

	return strings.Split(string(content), "\n"), nil
}
