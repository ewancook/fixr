package fixr

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

const decodeErrorMsg string = "1234" // for purposes of _decodeJSONResponseWrongErrorType()
var decodeTestJSON = []byte(fmt.Sprintf(`{"Error": "%s"}`, decodeErrorMsg))

func TestDecodeJSONResponseWrongErrorType(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewBuffer(decodeTestJSON))
	obj := &struct {
		Error float64 `json:",string"` // allow decoding
	}{}
	if err := decodeJSONResponse(body, obj); err != nil {
		t.Error(err)
	}
}

func TestDecodeJSONResponseShouldntDecode(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewBuffer(decodeTestJSON))
	obj := &struct{ Error float64 }{} // don't allow decoding
	if err := decodeJSONResponse(body, obj); err == nil || err.Error() == decodeErrorMsg {
		t.Error("JSON decoding should fail\n")
	}
}

func TestDecodeJSONResponseFinalReturnNil(t *testing.T) {
	emptyJSON := []byte(`{"Error": ""}`)
	body := ioutil.NopCloser(bytes.NewBuffer(emptyJSON))
	obj := &struct{ Error string }{}
	if err := decodeJSONResponse(body, obj); err != nil {
		t.Error(err)
	}
}

func TestDecodeJSONResponseNotAStruct(t *testing.T) {
	boolJSON := []byte(`true`)
	body := ioutil.NopCloser(bytes.NewBuffer(boolJSON))
	obj := false
	if err := decodeJSONResponse(body, &obj); err != nil {
		t.Error(err)
	}
}

func TestDecodeJSONResponseNormal(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewBuffer(decodeTestJSON))
	obj := &struct{ Error string }{"is this cleared?"}
	if err := decodeJSONResponse(body, obj).Error(); err != decodeErrorMsg {
		t.Errorf("expected '%s'; got '%s'\n", decodeErrorMsg, err)
	}
	if len(obj.Error) != 0 {
		t.Errorf("struct error field should be cleared. got: %+v", obj)
	}
}
