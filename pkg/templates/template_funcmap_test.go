package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCutFirstToLower(t *testing.T) {
	testCases := []struct {
		caseName string
		s        string
		result   string
	}{
		{caseName: "case1", s: "abc", result: "a"},
		{caseName: "case2", s: "我们", result: "我"},
		{caseName: "case3", s: "*+-\\'", result: "*"},
		{caseName: "case4", s: "_Acc", result: "_"},
		{caseName: "case5", s: "", result: ""},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.caseName, func(t *testing.T) {
			assert.Equal(t, test.result, CutFirstToLower(test.s))
		})
	}
}

func TestFirstToLower(t *testing.T) {
	testCases := []struct {
		caseName string
		s        string
		result   string
	}{
		{caseName: "First to lower", s: "Abc", result: "abc"},
		{caseName: "Chinese character", s: "我们", result: "我们"},
		{caseName: "Special symbols", s: "*+-\\'", result: "*+-\\'"},
		{caseName: "Empty string", s: "", result: ""},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.caseName, func(t *testing.T) {
			assert.Equal(t, test.result, FirstToLower(test.s))
		})
	}
}

func TestShortener(t *testing.T) {
	testCases := []struct {
		caseName string
		s        string
		result   string
	}{
		{caseName: "case1", s: "Abc", result: "a"},
		{caseName: "case2", s: "AbbCCaaDee", result: "acc"},
		{caseName: "case3", s: "_ab_i-u--n", result: "aiu"},
		{caseName: "case4", s: "TestShortener", result: "ts"},
		{caseName: "case5", s: "testshortener", result: "t"},
		{caseName: "case6", s: "test_shortener", result: "ts"},
		{caseName: "case7", s: "我们", result: "我"},
		{caseName: "Empty string", s: "", result: ""},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.caseName, func(t *testing.T) {
			assert.Equal(t, test.result, Shorten(test.s))
		})
	}
}

func TestFirstToUpper(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"string to up", args{"test"}, "Test"},
		{"num to up", args{"123"}, "123"},
		{"empty to up", args{""}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FirstToUpper(tt.args.s); got != tt.want {
				t.Errorf("FirstToUpper() = %v, want %v", got, tt.want)
			}
		})
	}
}
