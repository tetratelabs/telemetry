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

// Copyright (c) Tetrate, Inc 2022 All Rights Reserved.

package telemetry

import (
	"testing"
)

func TestFromLevel(t *testing.T) {
	tests := []struct {
		level string
		want  Level
		ok    bool
	}{
		{"none", LevelNone, true},
		{"error", LevelError, true},
		{"info", LevelInfo, true},
		{"debug", LevelDebug, true},
		{"invalid", LevelNone, false},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			level, ok := FromLevel(tt.level)

			if level != tt.want {
				t.Fatalf("AsLevel(%s)=%s, want: %s", tt.level, level, tt.want)
			}
			if ok != tt.ok {
				t.Fatalf("AsLevel(%s)=%t, want: %t", tt.level, ok, tt.ok)
			}
		})
	}
}
