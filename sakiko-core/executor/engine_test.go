package executor

import (
	"reflect"
	"testing"

	"sakiko.local/sakiko-core/interfaces"
)

func TestMacroExecutionOrderPrioritizesHeavyAndAuxiliaryMacrosLast(t *testing.T) {
	input := []interfaces.MacroType{
		interfaces.MacroMedia,
		interfaces.MacroSpeed,
		interfaces.MacroPing,
		interfaces.MacroGeo,
	}

	got := macroExecutionOrder(input)
	want := []interfaces.MacroType{
		interfaces.MacroPing,
		interfaces.MacroGeo,
		interfaces.MacroSpeed,
		interfaces.MacroMedia,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("macroExecutionOrder() = %v, want %v", got, want)
	}
}
