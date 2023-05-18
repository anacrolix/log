package log

import (
	"github.com/anacrolix/log/internal"
	qt "github.com/frankban/quicktest"
	"testing"
)

var globalVarCaller = getSingleCallerPc(0)

func globalFuncCaller() uintptr {
	return getSingleCallerPc(0)
}

type methodCaller struct{}

func (methodCaller) valueMethod() uintptr {
	return getSingleCallerPc(0)
}

func (*methodCaller) ptrMethod() uintptr {
	return getSingleCallerPc(0)
}

func checkPcPackage(c *qt.C, pc uintptr, expectedPkg string) {
	loc := locFromPc(pc)
	c.Log(loc.Function)
	c.Check(loc.Package, qt.Equals, expectedPkg)
}

func TestCallerLocs(t *testing.T) {
	c := qt.New(t)
	checkPcPackage(c, globalVarCaller, "github.com/anacrolix/log")
	checkPcPackage(c, globalFuncCaller(), "github.com/anacrolix/log")
	checkPcPackage(c, methodCaller{}.valueMethod(), "github.com/anacrolix/log")
	checkPcPackage(c, (*methodCaller).ptrMethod(nil), "github.com/anacrolix/log")
	var nestedPkgPc uintptr
	internal.Run(func() {
		nestedPkgPc = getSingleCallerPc(1)
	})
	checkPcPackage(c, nestedPkgPc, "github.com/anacrolix/log/internal")
}
