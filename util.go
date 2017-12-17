package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type ParamDecl map[string]reflect.Kind
type ParamValue map[string]reflect.Value

func getValue(kind reflect.Kind, input string) (*reflect.Value, error) {
	switch kind {
	case reflect.String:
		val := reflect.ValueOf(input)
		return &val, nil

	case reflect.Bool:
		b, err := strconv.ParseBool(input)

		if err != nil {
			return nil, err
		}

		val := reflect.ValueOf(b)
		return &val, nil

	case reflect.Int:
		i, err := strconv.ParseInt(input, 10, 64)

		if err != nil {
			return nil, err
		}

		val := reflect.ValueOf(i)
		return &val, nil

	default:
		return nil, fmt.Errorf("Can't handle param type: %s", kind)
	}
}

func ParamParse(req ParamDecl, args []string) (ParamValue, error) {
	var ret = make(map[string]reflect.Value)

	for _, i := range args {
		var elements []string = strings.SplitN(i, "=", 2)

		if len(elements) != 2 {
			return nil, fmt.Errorf("Invalid param %s", i)
		}

		var key string = strings.ToLower(strings.Trim(elements[0], " "))
		var val string = strings.Trim(elements[1], " ")

		param, ok := req[key]

		if !ok {
			return nil, fmt.Errorf("Don't know how to handle param: %s", i)
		}

		if _, exists := ret[key]; exists {
			return nil, fmt.Errorf("Don't specify multiple params: %s", key)
		}

		newVal, err := getValue(param, val)

		if err != nil {
			return nil, err
		}

		ret[key] = *newVal
	}

	return ret, nil
}
