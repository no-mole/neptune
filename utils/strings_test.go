package utils

import "testing"

func TestAToAbsI(t *testing.T) {
	expecteds := []*struct {
		Input       string
		expected    int64
		expectedErr bool
	}{
		{
			Input:       "a16",
			expected:    16,
			expectedErr: false,
		},
		{
			Input:       "16a",
			expected:    16,
			expectedErr: false,
		},
		{
			Input:       "a16a",
			expected:    16,
			expectedErr: false,
		},
		{
			Input:       "aa16aaa",
			expected:    16,
			expectedErr: false,
		},
		{
			Input:       "aaaa16aaa",
			expected:    16,
			expectedErr: false,
		},
		{
			Input:       "aaaa1a6aaa",
			expected:    0,
			expectedErr: true,
		},
		{
			Input:       "J039",
			expected:    39,
			expectedErr: false,
		},
		{
			Input:       "J0390",
			expected:    390,
			expectedErr: false,
		},
		{
			Input:       "J00390J",
			expected:    390,
			expectedErr: false,
		},
	}

	for _, item := range expecteds {
		n, err := AToAbsI(item.Input)
		if (item.expectedErr && err == nil) || n != item.expected {
			t.Errorf("input:%s,expected:%d,got:%d", item.Input, item.expected, n)
		}
	}
}
