# Telemetry

[![CI](https://github.com/tetratelabs/telemetry/actions/workflows/ci.yaml/badge.svg)](https://github.com/tetratelabs/telemetry/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/tetratelabs/telemetry/branch/master/graph/badge.svg?token=CLZMDX2TUN)](https://codecov.io/gh/tetratelabs/telemetry)

This package provides a set of Telemetry interfaces allowing you to completely
decouple your libraries and packages from Logging and Metrics instrumentation
implementations.

For more information on the interfaces, see: 
https://pkg.go.dev/github.com/tetratelabs/telemetry

## Implementations

Below you can find a list of known interface implementations. As a consumer of
this package, you might want to select existing implementations to use in your
application binaries. If looking to write your own, take a look at existing ones
for inspiration. 

If you have an OSS implementation you want to share, feel free to submit a PR
to this file so it may be included in this list. 

| repository | supported interfaces | notes |
| --- | --- | --- | 
| tetratelabs/[telemetry-gokit-log](https://github.com/tetratelabs/telemetry-gokit-log) | Logger | [Go kit log](https://github.com/go-kit/log) bridge |
| tetratelabs/[log](https://github.com/tetratelabs/log/tree/v2) | Logger | Scoped structured/unstructured logger bridge |
| tetratelabs/[telemetry-opencensus](https://github.com/tetratelabs/telemetry-opencensus) | Metrics | [OpenCensus metrics](https://github.com/census-instrumentation/opencensus-go) bridge |
| tetratelabs/[telemetry-opentelemetry](https://github.com/tetratelabs/telemetry-opentelemetry) | Metrics | [OpenTelemetry metrics](https://opentelemetry.io/) bridge |
