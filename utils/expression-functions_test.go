package utils

import (
	"fmt"
	"testing"
)

func TestIsoWeeFromIsoDate_shoulCacheLastUsedFormat(t *testing.T) {
	testCases := []struct {
		firstFormat    string
		secondFormat   string
		input          string
		expectedResult string
		hasError       bool
	}{
		{
			firstFormat:    supportedDateFormats[0],
			secondFormat:   supportedDateFormats[3],
			input:          "2026-07-15T18:43:29.293Z",
			expectedResult: "2026-29",
		},
		{
			firstFormat:    supportedDateFormats[2],
			secondFormat:   supportedDateFormats[0],
			input:          "20260721",
			expectedResult: "2026-30",
		},
		{
			firstFormat:    supportedDateFormats[0],
			secondFormat:   supportedDateFormats[3],
			input:          "2026-07-07T18:43:29.293452349-03:00",
			expectedResult: "2026-28",
		},
		{
			firstFormat:    supportedDateFormats[2],
			secondFormat:   supportedDateFormats[2],
			input:          "2026 07 07",
			expectedResult: "",
			hasError:       true,
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("input: %s | result: %s | error: %t", testCase.input, testCase.expectedResult, testCase.hasError), func(t *testing.T) {
			lastUsedDateStringFormat = testCase.firstFormat
			result, err := isoWeekFromIsoDate(testCase.input)
			if err != nil && testCase.hasError {
				return
			}
			if result != testCase.expectedResult {
				t.Errorf("expect %s to be equals to %s", result, testCase.expectedResult)
				return
			}
			if lastUsedDateStringFormat != testCase.secondFormat {
				t.Errorf("cache is not working (expect %s to be equal to %s)", lastUsedDateStringFormat, testCase.secondFormat)
				return
			}
		})
	}
}
