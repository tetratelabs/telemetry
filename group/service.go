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

// Package group provides a tetratelabs/run Group compatible scoped Logger
// configuration handler.
package group

import (
	"fmt"
	"strings"

	"github.com/tetratelabs/multierror"
	"github.com/tetratelabs/run"

	"github.com/tetratelabs/telemetry"
	"github.com/tetratelabs/telemetry/scope"
)

const (
	// LogOutputLevel is the name of the flag used to configure the logging levels.
	LogOutputLevel = "log-output-level"
)

const (
	// DefaultLogOutputLevel is the default level that will be configured for the loggers.
	DefaultLogOutputLevel = "info"
)

type service struct {
	outputLevels string
}

// New returns a new run Group Config to manage configuration of our scoped
// logger.
func New(l telemetry.Logger) run.Config {
	scope.UseLogger(l)
	return &service{}
}

// Name implements run.Unit.
func (s service) Name() string {
	return "log-manager"
}

// FlagSet implements run.Config.
func (s *service) FlagSet() *run.FlagSet {
	if s.outputLevels == "" {
		s.outputLevels = scope.DefaultLevel().String()
		if s.outputLevels == "" {
			s.outputLevels = DefaultLogOutputLevel
		}
	}
	fs := run.NewFlagSet("Logging options")
	fs.StringVar(&s.outputLevels, LogOutputLevel, s.outputLevels, fmt.Sprintf(
		"Comma-separated minimum per-scope logging level of messages to output, "+
			"in the form of [default_level,]<scope>:<level>,<scope>:<level>,... "+
			"where scope can be one of [%s] and default_level or level can be "+
			"one of [%s]",
		strings.Join(scope.Names(), ", "),
		strings.Join([]string{"debug", "info", "error", "none"}, ", "),
	))

	return fs
}

// Validate implements run.Config.
func (s *service) Validate() error {
	var mErr error

	s.outputLevels = strings.ToLower(s.outputLevels)
	outputLevels := strings.Split(s.outputLevels, ",")
	if len(outputLevels) == 0 {
		return nil
	}
	for _, ol := range outputLevels {
		osl := strings.Split(ol, ":")
		switch len(osl) {
		case 1:
			lvl, ok := telemetry.FromLevel(strings.Trim(ol, "\r\n\t "))
			if !ok {
				mErr = multierror.Append(mErr, fmt.Errorf("%q is not a valid log level", ol))
				continue
			}
			scope.SetAllScopes(lvl)
		case 2:
			lvl, ok := telemetry.FromLevel(strings.Trim(osl[1], "\r\n\t "))
			if !ok {
				mErr = multierror.Append(mErr, fmt.Errorf("%q is not a valid log level", ol))
				continue
			}
			if s := scope.Find(osl[0]); s != nil {
				s.SetLevel(lvl)
			} else {
				mErr = multierror.Append(mErr, fmt.Errorf("%q is not a registered scope", osl[0]))
			}
		default:
			mErr = multierror.Append(mErr, fmt.Errorf("%q is not a valid <scope>:<level> pair", ol))
		}
	}

	return mErr
}
