package log

import (
	"testing"

	"github.com/go-quicktest/qt"

	"github.com/anacrolix/log/internal"
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

func checkPcPackage(t testing.TB, pc uintptr, expectedPkg string) {
	loc := locFromPc(pc)
	t.Log(loc.Function)
	qt.Check(t, qt.Equals(loc.Package, expectedPkg))
}

func TestCallerLocs(t *testing.T) {
	checkPcPackage(t, globalVarCaller, "github.com/anacrolix/log")
	checkPcPackage(t, globalFuncCaller(), "github.com/anacrolix/log")
	checkPcPackage(t, methodCaller{}.valueMethod(), "github.com/anacrolix/log")
	checkPcPackage(t, (*methodCaller).ptrMethod(nil), "github.com/anacrolix/log")
	var nestedPkgPc uintptr
	internal.Run(func() {
		nestedPkgPc = getSingleCallerPc(1)
	})
	checkPcPackage(t, nestedPkgPc, "github.com/anacrolix/log/internal")
}
