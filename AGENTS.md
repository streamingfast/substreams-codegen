# Golang project

## Tests

- Highly favor table driven test cases
- Use "github.com/stretchr/testify/assert" & "github.com/stretchr/testify/require" for all assertion needs
- Use `require.NoError(t, err)` to all error checking
- Each testing helper that receives `t` as first argument should call `t.Helper()`