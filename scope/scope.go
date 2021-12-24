// Copyright (c) Tetrate, Inc 2021.
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

// Package scope provides a scoped logger facade for telemetry.Logger
// implementations.
package scope

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/tetratelabs/telemetry"
	"github.com/tetratelabs/telemetry/level"
)

var (
	_ level.Logger = (*scope)(nil)

	lock          = sync.Mutex{}
	scopes        = make(map[string]*scope)
	uninitialized = make(map[string][]*scope)
	defaultLevel  = Info
	defaultLogger level.Logger

	// PanicOnUninitialized can be used when testing for sequencing issues
	// between creating log lines and initializing the actual logger
	// implementation to use.
	PanicOnUninitialized bool
)

// Available log levels.
const (
	None  = level.None
	Error = level.Error
	Info  = level.Info
	Debug = level.Debug
)

var levelToString = map[level.Value]string{
	None:  "none",
	Error: "error",
	Info:  "info",
	Debug: "debug",
}

// scope provides scoped logging functionality.
type scope struct {
	logger      telemetry.Logger
	kvs         []interface{}
	ctx         context.Context
	metric      telemetry.Metric
	name        string
	description string
}

// Debug implements telemetry.Logger.
func (s *scope) Debug(msg string, keyValuePairs ...interface{}) {
	if s.logger != nil {
		s.logger.Debug(msg, keyValuePairs)
	}
	if PanicOnUninitialized {
		panic("calling Debug on uninitialized logger")
	}
}

// Info implements telemetry.Logger.
func (s *scope) Info(msg string, keyValuePairs ...interface{}) {
	if s.logger != nil {
		s.logger.Debug(msg, keyValuePairs)
	}
	if PanicOnUninitialized {
		panic("calling Info on uninitialized logger")
	}
}

// Error implements telemetry.Logger.
func (s *scope) Error(msg string, err error, keyValuePairs ...interface{}) {
	if s.logger != nil {
		s.logger.Error(msg, err, keyValuePairs...)
	}
	if PanicOnUninitialized {
		panic("calling Error on uninitialized logger")
	}
}

// With implements telemetry.Logger.
func (s *scope) With(keyValuePairs ...interface{}) telemetry.Logger {
	if len(keyValuePairs) == 0 {
		return s
	}
	if len(keyValuePairs)%2 != 0 {
		keyValuePairs = append(keyValuePairs, "(MISSING)")
	}
	if s.logger != nil {
		return s.logger.With(keyValuePairs)
	}
	sc := &scope{
		name:        s.name,
		description: s.description,
		kvs:         make([]interface{}, len(s.kvs), len(s.kvs)+len(keyValuePairs)),
		ctx:         s.ctx,
		metric:      s.metric,
	}
	copy(sc.kvs, keyValuePairs)
	for i := 0; i < len(keyValuePairs); i += 2 {
		if k, ok := keyValuePairs[i].(string); ok {
			sc.kvs = append(sc.kvs, k, keyValuePairs[i+1])
		}
	}
	uninitialized[s.name] = append(uninitialized[s.name], sc)

	return sc
}

// Context implements telemetry.Logger.
func (s *scope) Context(ctx context.Context) telemetry.Logger {
	if s.logger != nil {
		return s.logger.Context(ctx)
	}
	sc := &scope{
		name:        s.name,
		description: s.description,
		kvs:         make([]interface{}, len(s.kvs)),
		ctx:         ctx,
		metric:      s.metric,
	}
	copy(sc.kvs, s.kvs)
	uninitialized[s.name] = append(uninitialized[s.name], sc)

	return sc
}

// Metric implements telemetry.Logger.
func (s *scope) Metric(m telemetry.Metric) telemetry.Logger {
	if s.logger != nil {
		return s.logger.Metric(m)
	}
	scope := &scope{
		name:        s.name,
		description: s.description,
		kvs:         make([]interface{}, len(s.kvs)),
		ctx:         s.ctx,
		metric:      s.metric,
	}
	copy(scope.kvs, s.kvs)
	uninitialized[s.name] = append(uninitialized[s.name], scope)

	return scope
}

// New implements level.Logger.
func (s *scope) New() telemetry.Logger {
	if s.logger != nil {
		return s.logger.(level.Logger).New()
	}
	return s
}

// SetLevel implements level.Logger.
func (s *scope) SetLevel(lvl level.Value) {
	if s.logger != nil {
		s.logger.(level.Logger).SetLevel(lvl)
	}
}

// Level implements level.Logger.
func (s *scope) Level() level.Value {
	if s.logger != nil {
		return s.logger.(level.Logger).Level()
	}
	return level.None
}

// Register a new scoped Logger.
func Register(name, description string) level.Logger {
	if strings.ContainsAny(name, ":,.") {
		return nil
	}

	lock.Lock()
	defer lock.Unlock()

	name = strings.ToLower(strings.Trim(name, "\r\n\t "))
	sc, ok := scopes[name]
	if !ok {
		sc = &scope{
			name:        name,
			description: description,
		}
		if defaultLogger != nil {
			sc.logger = defaultLogger.New()
		}
		scopes[name] = sc
	}
	if defaultLogger == nil {
		uninitialized[name] = append(uninitialized[name], sc)
	}

	return sc
}

// Find a scoped logger by its name.
func Find(name string) level.Logger {
	lock.Lock()
	defer lock.Unlock()

	name = strings.ToLower(strings.Trim(name, "\r\n\t "))
	return scopes[name]
}

// List all registered Scopes
func List() map[string]level.Logger {
	lock.Lock()
	defer lock.Unlock()

	sc := make(map[string]level.Logger, len(scopes))
	for k, v := range scopes {
		sc[k] = v
	}

	return sc
}

// Names returns all registered scope names.
func Names() []string {
	lock.Lock()
	defer lock.Unlock()

	var s []string
	for k := range scopes {
		s = append(s, k)
	}

	return s
}

func PrintRegistered() {
	lock.Lock()
	defer lock.Unlock()

	pad := 7
	names := make([]string, 0, len(scopes))
	for n := range scopes {
		names = append(names, n)
		if len(n) > pad {
			pad = len(n)
		}
	}
	sort.Strings(names)

	fmt.Println("registered logging scopes:")
	fmt.Printf("- %-*s [%-5s]  %s\n",
		pad,
		"default",
		levelToString[defaultLevel],
		"",
	)
	for _, n := range names {
		sc := scopes[n]
		fmt.Printf("- %-*s [%-5s]  %s\n",
			pad,
			sc.name,
			levelToString[sc.Level()],
			sc.description,
		)
	}
}

// SetDefaultLevel sets the default level used for new scopes.
func SetDefaultLevel(lvl level.Value) {
	lock.Lock()
	defer lock.Unlock()

	if defaultLogger != nil {
		defaultLogger.SetLevel(lvl)
	}
}

// DefaultLevel returns the logging level used for new scopes.
func DefaultLevel() level.Value {
	if defaultLogger != nil {
		return defaultLogger.Level()
	}
	return level.None
}

// UseLogger takes a logger and updates already registered scopes to use it.
// This function can only be used once. It can't override an already initialized
// logger. Therefore, set this as soon as possible.
func UseLogger(logger telemetry.Logger) {
	if logger == nil {
		return
	}

	lock.Lock()
	defer lock.Unlock()

	if defaultLogger != nil {
		return
	}

	// update our default logger
	// if provided Logger does not provide log level logic, wrap it.
	var ok bool
	defaultLogger, ok = logger.(level.Logger)
	if !ok {
		defaultLogger = level.Wrap(logger)
	}

	// adjust already registered scopes
	for _, scopes := range uninitialized {
		l := defaultLogger.New()
		for _, sc := range scopes {
			if sc.ctx != nil {
				l = l.Context(sc.ctx)
			}
			if sc.metric != nil {
				l = l.Metric(sc.metric)
			}
			if len(sc.kvs) > 0 {
				l = l.With(sc.kvs...)
			}
			sc.logger = l
			sc.kvs = nil
			sc.ctx = nil
			sc.metric = nil
		}
	}
	uninitialized = nil
}
