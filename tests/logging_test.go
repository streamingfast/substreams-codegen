package tests

import "github.com/streamingfast/logging"

var zlog, _ = logging.PackageLogger("test", "github.com/streamingfast/substreams-codegen/tests")

func init() {
	logging.InstantiateLoggers()
}
