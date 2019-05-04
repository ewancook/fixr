package fixr

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func _decodeJSONResponseWrongErrorType(t *testing.T, testJSON []byte) {
	body := ioutil.NopCloser(bytes.NewBuffer(testJSON))
	obj := &struct {
		Error float64 `json:",string"` // allow decoding
	}{}
	if err := decodeJSONResponse(body, obj); err != nil {
		t.Error(err)
	}
}

func _decodeJSONResponseShouldntDecode(t *testing.T, errorMessage string, testJSON []byte) {
	body := ioutil.NopCloser(bytes.NewBuffer(testJSON))
	obj := &struct{ Error float64 }{} // don't allow decoding
	if err := decodeJSONResponse(body, obj); err == nil || err.Error() == errorMessage {
		t.Error("JSON decoding should fail\n")
	}
}

func _decodeJSONResponseNormal(t *testing.T, errorMessage string, testJSON []byte) {
	body := ioutil.NopCloser(bytes.NewBuffer(testJSON))
	obj := &struct{ Error string }{"is this cleared?"}
	if err := decodeJSONResponse(body, obj).Error(); err != errorMessage {
		t.Errorf("expected '%s'; got '%s'\n", errorMessage, err)
	}
	if len(obj.Error) != 0 {
		t.Errorf("struct error field should be cleared. got: %+v", obj)
	}
}

func TestDecodeJSONResponse(t *testing.T) {
	errorMessage := "1234" // for purposes of _decodeJSONResponseWrongErrorType()
	testJSON := []byte(fmt.Sprintf(`{"Error": "%s"}`, errorMessage))
	_decodeJSONResponseNormal(t, errorMessage, testJSON)
	_decodeJSONResponseShouldntDecode(t, errorMessage, testJSON)
	_decodeJSONResponseWrongErrorType(t, testJSON)
}
