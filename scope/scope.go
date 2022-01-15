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

// Package scope provides a scoped logger facade for telemetry.Logger
// implementations.
package scope

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/tetratelabs/telemetry"
)

var (
	_ telemetry.Logger = (*scope)(nil)

	lock          = sync.Mutex{}
	scopes        = make(map[string]*scope)
	uninitialized = make(map[string][]*scope)
	defaultLogger telemetry.Logger

	// PanicOnUninitialized can be used when testing for sequencing issues
	// between creating log lines and initializing the actual logger
	// implementation to use.
	PanicOnUninitialized bool
)

const (
	// Key used to store the name of scope in the logger key/value pairs.
	Key = "scope"
)

// scope provides scoped logging functionality.
type scope struct {
	logger      telemetry.Logger
	kvs         []interface{}
	ctx         context.Context
	metric      telemetry.Metric
	name        string
	description string
	level       *int32
}

// Debug implements telemetry.Logger.
func (s *scope) Debug(msg string, keyValuePairs ...interface{}) {
	if s.logger != nil {
		s.logger.Debug(msg, keyValuePairs...)
	}
	if PanicOnUninitialized {
		panic("calling Debug on uninitialized logger")
	}
}

// Info implements telemetry.Logger.
func (s *scope) Info(msg string, keyValuePairs ...interface{}) {
	if s.logger != nil {
		s.logger.Info(msg, keyValuePairs...)
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
		return s.logger.With(keyValuePairs...)
	}
	sc := &scope{
		name:        s.name,
		description: s.description,
		kvs:         make([]interface{}, len(s.kvs), len(s.kvs)+len(keyValuePairs)),
		ctx:         s.ctx,
		metric:      s.metric,
		level:       s.level,
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

	sc := s.Clone()
	sc.(*scope).ctx = ctx
	uninitialized[s.name] = append(uninitialized[s.name], sc.(*scope))
	return sc
}

// Metric implements telemetry.Logger.
func (s *scope) Metric(m telemetry.Metric) telemetry.Logger {
	if s.logger != nil {
		return s.logger.Metric(m)
	}

	sc := s.Clone()
	sc.(*scope).metric = m
	uninitialized[s.name] = append(uninitialized[s.name], sc.(*scope))
	return sc
}

// Clone implements level.Logger.
func (s *scope) Clone() telemetry.Logger {
	var logger telemetry.Logger
	if s.logger != nil {
		logger = s.logger.Clone()
	}

	scope := &scope{
		logger:      logger,
		name:        s.name,
		description: s.description,
		kvs:         make([]interface{}, len(s.kvs)),
		ctx:         s.ctx,
		metric:      s.metric,
		level:       s.level,
	}

	copy(scope.kvs, s.kvs)

	return scope
}

// SetLevel implements level.Logger.
func (s *scope) SetLevel(lvl telemetry.Level) {
	if s.logger != nil {
		s.logger.SetLevel(lvl)
		return
	}

	switch {
	case lvl < telemetry.LevelError:
		lvl = telemetry.LevelNone
	case lvl < telemetry.LevelInfo:
		lvl = telemetry.LevelError
	case lvl < telemetry.LevelDebug:
		lvl = telemetry.LevelInfo
	default:
		lvl = telemetry.LevelDebug
	}

	atomic.StoreInt32(s.level, int32(lvl))
}

// Level implements level.Logger.
func (s *scope) Level() telemetry.Level {
	if s.logger != nil {
		return s.logger.Level()
	}
	return telemetry.Level(atomic.LoadInt32(s.level))
}

// Register a new scoped Logger.
func Register(name, description string) telemetry.Logger {
	if strings.ContainsAny(name, ":,.") {
		return nil
	}

	lock.Lock()
	defer lock.Unlock()

	name = strings.ToLower(strings.Trim(name, "\r\n\t "))
	sc, ok := scopes[name]
	if !ok {
		level := int32(DefaultLevel())
		sc = &scope{
			name:        name,
			description: description,
			ctx:         context.Background(),
			kvs:         []interface{}{Key, name},
			level:       &level,
		}
		if defaultLogger != nil {
			sc.logger = defaultLogger.With(Key, name)
		}

		scopes[name] = sc
	}
	if defaultLogger == nil {
		uninitialized[name] = append(uninitialized[name], sc)
	}

	return sc
}

// Find a scoped logger by its name.
func Find(name string) (telemetry.Logger, bool) {
	lock.Lock()
	defer lock.Unlock()

	name = strings.ToLower(strings.Trim(name, "\r\n\t "))
	s, ok := scopes[name]
	return s, ok
}

// List all registered Scopes
func List() map[string]telemetry.Logger {
	lock.Lock()
	defer lock.Unlock()

	sc := make(map[string]telemetry.Logger, len(scopes))
	for k, v := range scopes {
		sc[k] = v
	}

	return sc
}

// Names returns all registered scope names.
func Names() []string {
	lock.Lock()
	defer lock.Unlock()

	s := make([]string, 0, len(scopes))
	for k := range scopes {
		s = append(s, k)
	}

	return s
}

// PrintRegistered outputs a list of registered scopes with their log level on
// stdout.
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
		DefaultLevel().String(),
		"",
	)
	for _, n := range names {
		sc := scopes[n]
		fmt.Printf("- %-*s [%-5s]  %s\n",
			pad,
			sc.name,
			sc.Level().String(),
			sc.description,
		)
	}
}

// SetAllScopes sets the logging level to all existing scopes and uses this
// level for new scopes.
func SetAllScopes(lvl telemetry.Level) {
	lock.Lock()
	defer lock.Unlock()

	if defaultLogger != nil {
		defaultLogger.SetLevel(lvl)
		for _, sc := range scopes {
			sc.SetLevel(lvl)
		}
	}
}

// SetDefaultLevel sets the default level used for new scopes.
func SetDefaultLevel(lvl telemetry.Level) {
	lock.Lock()
	defer lock.Unlock()

	if defaultLogger != nil {
		defaultLogger.SetLevel(lvl)
	}
}

// DefaultLevel returns the logging level used for new scopes.
func DefaultLevel() telemetry.Level {
	if defaultLogger != nil {
		return defaultLogger.Level()
	}
	return telemetry.LevelNone
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

	defaultLogger = logger.Clone()

	// adjust already registered scopes
	for _, scopes := range uninitialized {
		l := defaultLogger.Clone()
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
			l.SetLevel(sc.Level())

			sc.logger = l
			sc.kvs = nil
			sc.ctx = nil
			sc.metric = nil
			sc.level = nil
		}
	}
	uninitialized = nil
}
