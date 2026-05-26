package utils

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/wisemonkeys-co/goval"
)

func MapFunctions() (functions map[string]goval.ExpressionFunction) {
	functions = make(map[string]goval.ExpressionFunction)
	functions["getRangeValue"] = getRangeValue
	functions["strToInt"] = strToInt
	functions["min"] = min
	functions["getItemFromStringArray"] = getItemFromStringArray
	functions["parseFloat"] = parseFloat
	return
}

func getRangeValue(args ...any) (any, error) {
	var defaultValue any
	defaultValue = 0.0
	if len(args) > 3 {
		defaultValue = args[3].(float64)
	}
	amount, err := getInteger(args[0])
	if err != nil {
		return defaultValue, fmt.Errorf("could not convert %v (%T) to an integer (amount) [%s]", args[0], args[0], err.Error())
	}
	ranges := args[1].(map[string]any)
	field := "value"
	if len(args) > 2 {
		field = args[2].(string)
	}
	for _, v := range ranges {
		dataMap, ok := v.(map[string]any)
		if !ok {
			return defaultValue, fmt.Errorf("could not convert %v (%T) to a valid data map", v, v)
		}
		min, err := getInteger(dataMap["min"])
		if err != nil {
			return defaultValue, fmt.Errorf("could not convert %v (%T) to an integer (min) [%s]", dataMap["min"], dataMap["min"], err.Error())
		}
		max, err := getInteger(dataMap["max"])
		if err != nil {
			return defaultValue, fmt.Errorf("could not convert %v (%T) to an integer (max) [%s]", dataMap["max"], dataMap["max"], err.Error())
		}
		if amount <= max && amount >= min {
			return dataMap[field], nil
		}
	}
	return defaultValue, nil
}

func strToInt(args ...any) (any, error) {
	return strconv.Atoi(args[0].(string))
}

func min(args ...any) (any, error) {
	v1, err := getFloat(args[0])
	if err != nil {
		return v1, err
	}
	v2, err := getFloat(args[1])
	if err != nil {
		return v2, err
	}
	return math.Min(v1, v2), nil
}

func getItemFromStringArray(args ...any) (data any, err error) {
	array := args[0].([]string)
	index, err := getInteger(args[1])
	if err != nil {
		return
	}
	data = array[index]
	return
}

func parseFloat(args ...any) (data any, err error) {
	switch i := args[0].(type) {
	case float64:
		return i, nil
	case int:
		return float64(i), nil
	case string:
		return strconv.ParseFloat(i, 64)
	default:
		return math.NaN(), errors.New("non-numeric type could not be converted to float")
	}
}

func getFloat(unk any) (float64, error) {
	switch i := unk.(type) {
	case float64:
		return i, nil
	case int:
		return float64(i), nil
	default:
		return math.NaN(), errors.New("non-numeric type could not be converted to float")
	}
}

func getInteger(unk any) (int, error) {
	switch i := unk.(type) {
	case float64:
		return int(i), nil
	case int:
		return i, nil
	case int32:
		return int(i), nil
	case int64:
		return int(i), nil
	default:
		return 0, errors.New("non-numeric type could not be converted to float")
	}
}
