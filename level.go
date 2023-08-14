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

// Level is an enumeration of the available log levels.
type Level int32

// Available log levels.
const (
	LevelNone  Level = 0
	LevelError Level = 1
	LevelInfo  Level = 5
	LevelDebug Level = 10
)

// levelToString maps each logging level to its string representation.
var levelToString = map[Level]string{
	LevelNone:  "none",
	LevelError: "error",
	LevelInfo:  "info",
	LevelDebug: "debug",
}

// levelToString maps string representations to the corresponding level
var stringToLevel = map[string]Level{
	"none":  LevelNone,
	"error": LevelError,
	"info":  LevelInfo,
	"debug": LevelDebug,
}

// String returns the string representation of the logging level.
func (v Level) String() string { return levelToString[v] }

// FromLevel returns the logging level corresponding to the given string representation.
func FromLevel(level string) (Level, bool) {
	l, ok := stringToLevel[level]
	return l, ok
}
