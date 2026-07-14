package helpers_test

import (
	"quick-crud/helpers"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSnakeCase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
	}{
		{in: "test", want: "test"},
		{in: "Test", want: "test"},
		{in: "TestOne", want: "test_one"},
		{in: "testOne", want: "test_one"},
		{in: "User5", want: "user5"},
	}
	for _, tt := range tests {
		got := helpers.ToSnakeCase(tt.in)
		assert.Equal(t, tt.want, got)
	}
}
