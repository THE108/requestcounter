package params

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Params is request params
type Params struct {
	Request *http.Request
}

// NewParams returns new request params
func NewParams(request *http.Request) Params {
	return Params{
		Request: request,
	}
}

func getErrorIfRequiredMissed(paramName string, required bool) error {
	if required {
		return fmt.Errorf("required param %s is missing", paramName)
	}
	return nil
}

// String returns param value if exists in query string
// returns error if one of param not exists
func (params *Params) String(key string, required bool, defaultValue ...string) (string, error) {
	var defaultVal string
	if len(defaultValue) > 0 {
		defaultVal = defaultValue[0]
	}

	vars := mux.Vars(params.Request)
	if vars == nil {
		return defaultVal, getErrorIfRequiredMissed(key, required)
	}

	if value, ok := vars[key]; ok {
		return value, nil
	}

	if values, ok := params.Request.URL.Query()[key]; ok {
		return values[0], nil
	}

	return defaultVal, getErrorIfRequiredMissed(key, required)
}

// Bool returns bool-value from query string
func (params *Params) Bool(key string, required bool, defaultValue ...bool) (bool, error) {
	var defaultVal bool
	if len(defaultValue) > 0 {
		defaultVal = defaultValue[0]
	}

	strVal, err := params.String(key, required)
	if err != nil {
		return defaultVal, err
	}

	value, err := strconv.ParseBool(strVal)
	if err != nil {
		return defaultVal, fmt.Errorf("invalid boolean %s specified:%q error:%s", key, strVal, err)
	}

	return value, nil
}

// Uint returns uint64-value from query string
func (params *Params) Uint64(key string, required bool, defaultValue ...uint64) (uint64, error) {
	var defaultVal uint64
	if len(defaultValue) > 0 {
		defaultVal = defaultValue[0]
	}

	strVal, err := params.String(key, required)
	if err != nil {
		return defaultVal, err
	}

	value, err := strconv.ParseUint(strVal, 10, 64)
	if err != nil {
		return defaultVal, fmt.Errorf("invalid integer %s specified:%q error:%s", key, strVal, err)
	}

	return value, nil
}
