package ethfull

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeStructABI(t *testing.T) {
	cases := []struct {
		name               string
		inputStructAbiName string
		expectedName       string
	}{
		{
			inputStructAbiName: "foo_bar",
			expectedName:       "foo_bar",
		},
		{
			inputStructAbiName: "foo___bar",
			expectedName:       "foo_u_u_bar",
		},
		{
			inputStructAbiName: "__foo_bar",
			expectedName:       "u_u_foo_bar",
		},
		{
			inputStructAbiName: "foobar__",
			expectedName:       "foobar_u_",
		},
		{
			inputStructAbiName: "__foobar",
			expectedName:       "u_u_foobar",
		},
	}

	for _, c := range cases {
		t.Run(c.inputStructAbiName, func(t *testing.T) {
			result := sanitizeABIStructName(c.inputStructAbiName)
			require.Equal(t, c.expectedName, result)
		})
	}
}
