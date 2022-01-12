module github.com/tetratelabs/telemetry/group

go 1.17

require (
	github.com/tetratelabs/multierror v1.1.0
	github.com/tetratelabs/run v0.1.0
	github.com/tetratelabs/telemetry v0.5.2
)

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

// Work around for maintaining multiple go modules in the same repository
// until go has better support for this. https://github.com/golang/go/issues/45713
replace github.com/tetratelabs/telemetry => ../
