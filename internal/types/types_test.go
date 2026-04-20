package types

import (
	"reflect"
	"testing"
)

func TestHomeDataTodayTaskUsesPointerType(t *testing.T) {
	t.Parallel()

	field, ok := reflect.TypeOf(HomeData{}).FieldByName("TodayTask")
	if !ok {
		t.Fatal("TodayTask field not found")
	}
	if field.Type.Kind() != reflect.Pointer {
		t.Fatalf("expected TodayTask to be a pointer field, got %s", field.Type.Kind())
	}
}
