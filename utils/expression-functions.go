package utils

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/wisemonkeys-co/goval"
)

func MapFunctions() (functions map[string]goval.ExpressionFunction) {
	functions = make(map[string]goval.ExpressionFunction)
	functions["getRangeValue"] = getRangeValue
	functions["strToInt"] = strToInt
	functions["min"] = min
	functions["parseFloat"] = parseFloat
	functions["isoWeekFromIsoDate"] = isoWeekFromIsoDate
	functions["isoWeekFromYearMothDayStrings"] = isoWeekFromYearMothDayStrings
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

// isoWeekFromIsoDate:
// receives the iso date
// returns the iso week
// Example:
// consider event.date = "2026-06-23T14:30:00Z"
// isoWeekFromIsoDate(event.date)
func isoWeekFromIsoDate(args ...any) (any, error) {
	dateStr := args[0].(string)
	date, err := time.ParseInLocation(time.RFC3339, dateStr, nil)
	if err != nil {
		return "", err
	}
	year, week := date.ISOWeek()
	return fmt.Sprintf("%d-%d", year, week), nil
}

// isoWeekFromYearMothDayStrings: receives the year, month, and date as strings
// Return the iso week
// Example:
// consider event.date = "2026-06-23T14:30:00Z"
// isoWeekFromStrs(event.date[0:4], event.date[5:7], event.date[8:10])
func isoWeekFromYearMothDayStrings(args ...any) (any, error) {
	if len(args) < 3 {
		return "", fmt.Errorf("expect year, month and day as params but received %d params", len(args))
	}
	yearStr := args[0].(string)
	monthStr := args[1].(string)
	dayStr := args[2].(string)
	date, err := time.ParseInLocation("20060102", yearStr+monthStr+dayStr, nil)
	if err != nil {
		return "", err
	}
	year, week := date.ISOWeek()
	return fmt.Sprintf("%d-%d", year, week), nil
}
