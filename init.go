package twelve

import "github.com/hootuu/utils/logger"

var gLogger = logger.GetLogger("twelve")

var gExpectFactory = NewExpectFactory()

func init() {
	gExpectFactory.StartGC()
}
