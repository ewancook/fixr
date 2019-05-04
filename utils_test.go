package fixr

import (
	"fmt"
	"math/rand"
	"testing"
)

func init() {
	seededRand = rand.New(rand.NewSource(0))
}

func TestUUID(t *testing.T) {
	result, expected := uuid(), "f1f8"
	if result != expected {
		t.Errorf("expected %s; got %s\n", expected, result)
	}
}

func TestGenKey(t *testing.T) {
	result, expected := genKey(), "3eb6a7ec-0de9-5e1a-4a1b-3143a7c3e5ac"
	if result != expected {
		t.Errorf("expected %s; got %s\n", expected, result)
	}
}

func TestUnmarshalOutputNormal(t *testing.T) {
	expected := "1.0"
	testJSON := fmt.Sprintf(`{"APP_VERSION": "%s"}`, expected)
	result, _ := unmarshalOutput(testJSON)
	if result.Version != expected {
		t.Errorf("expected %s; got %s\n", expected, result.Version)
	}
}

func TestUnmarshalOutputUnmarshalFailure(t *testing.T) {
	if _, err := unmarshalOutput(`{test}`); err == nil {
		t.Error("JSON unmarshalling should fail")
	}
}

func TestUnmarshalOutputNotJSONArray(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected: runtime error: slice bounds out of range")
		}
	}()
	unmarshalOutput(``)
}
