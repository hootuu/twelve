package twelve

import "github.com/hootuu/utils/logger"

var gLogger = logger.GetLogger("twelve", logger.Options{Dir: "twelve"})
var gChainLogger = logger.GetLogger("chain", logger.Options{Dir: "twelve"})

var gExpectFactory = NewExpectFactory()

func init() {
	gExpectFactory.StartGC()
}
