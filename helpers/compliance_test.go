package helpers

import (
	"reflect"
	"testing"
)

func TestFacadeImplementsEveryInterfaceMethod(t *testing.T) {
	iface := reflect.TypeOf((*HelperInterface)(nil)).Elem()
	impl := reflect.TypeOf((*Helpers)(nil))
	for i := 0; i < iface.NumMethod(); i++ {
		m := iface.Method(i)
		if _, ok := impl.MethodByName(m.Name); !ok {
			t.Errorf("Helpers missing %s", m.Name)
		}
	}
	if iface.NumMethod() < 80 {
		t.Fatalf("interface shrank unexpectedly: %d", iface.NumMethod())
	}
}
