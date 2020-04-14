package templates

import (
	"testing"
	"github.com/magiconair/properties/assert"
)


func TestSnakeToCamel(t *testing.T)  {
	testCases := []struct {
		caseName string
		s string
		result string
	}{
		{caseName:  "case1" , s : "abc", result: "Abc"},
		{caseName:  "case2" , s : "a_b_c", result: "ABC"},
		{caseName:  "case3" , s : "abc_bcd_cde", result: "AbcBcdCde"},
	}

	for _, test := range testCases{
		t.Run(test.caseName, func(t *testing.T) {
			assert.Equal(t, test.result, SnakeToCamel(test.s))

		})
	}
}


func TestCutFirstToLower(t *testing.T)  {

	testCases := []struct {
		caseName string
		s string
		result string
	}{
		{caseName:  "case1" , s : "abc", result: "a"},
		{caseName:  "case2" , s : "我们", result: "我"},
		{caseName:  "case2" , s : "*+-\\'", result: "*"},
	}

	for _, test := range testCases{
		t.Run(test.caseName, func(t *testing.T) {
			assert.Equal(t, test.result, CutFirstToLower(test.s))

		})
	}
}




func TestFirstToLower(t *testing.T)  {

	testCases := []struct {
		caseName string
		s string
		result string
	}{
		{caseName:  "case1" , s : "Abc", result: "abc"},
		{caseName:  "case2" , s : "我们", result: "我们"},
		{caseName:  "case2" , s : "*+-\\'", result: "*+-\\'"},
	}

	for _, test := range testCases{
		t.Run(test.caseName, func(t *testing.T) {
			assert.Equal(t, test.result, FirstToLower(test.s))

		})
	}
}

func TestCamelToSnake(t *testing.T)  {
	testCases := []struct {
		caseName string
		s string
		result string
	}{
		{caseName:  "case1" , s : "Abc", result: "abc"},
		{caseName:  "case2" , s : "ABC", result: "a_b_c"},
		{caseName:  "case3" , s : "AbcBcdCde", result: "abc_bcd_cde"},
		{caseName:  "case4" , s : "", result: ""},
	}

	for _, test := range testCases{
		t.Run(test.caseName, func(t *testing.T) {
			assert.Equal(t, test.result, CamelToSnake(test.s))

		})
	}
}