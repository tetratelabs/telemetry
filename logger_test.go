// Copyright (c) Tetrate, Inc 2023.
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

package telemetry

import (
	"context"
	"reflect"
	"testing"
)

func TestContext(t *testing.T) {
	want := []interface{}{"key1", "val1"}
	ctx := context.Background()

	ctx = KeyValuesToContext(ctx, want...)
	have := KeyValuesFromContext(ctx)

	if !reflect.DeepEqual(want, have) {
		t.Errorf("want: %+v\nhave: %+v\n", want, have)
	}
}

func TestRemoveFromContext(t *testing.T) {
	want := []interface{}{"key1", "val2"}
	ctx := context.Background()

	ctx = KeyValuesToContext(ctx, "key1", "val1")
	ctx = RemoveKeyValuesFromContext(ctx)
	ctx = KeyValuesToContext(ctx, want...)
	have := KeyValuesFromContext(ctx)

	if !reflect.DeepEqual(want, have) {
		t.Errorf("want: %+v\nhave: %+v\n", want, have)
	}
}
