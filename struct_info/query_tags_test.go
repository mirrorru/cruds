package struct_info_test

import (
	"testing"

	"github.com/mirrorru/crudquick/struct_info"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseQueryTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tag      string
		expected struct_info.QueryTagFlags
	}{
		{
			name:     "empty tag",
			tag:      "",
			expected: struct_info.QueryTagFlags{},
		},
		{
			name: "from only",
			tag:  "from",
			expected: struct_info.QueryTagFlags{
				IsFrom: true,
			},
		},
		{
			name: "join=left",
			tag:  "join=left",
			expected: struct_info.QueryTagFlags{
				JoinType: "left",
			},
		},
		{
			name: "join=RIGHT (uppercase)",
			tag:  "join=RIGHT",
			expected: struct_info.QueryTagFlags{
				JoinType: "right",
			},
		},
		{
			name: "alias",
			tag:  "alias=u1",
			expected: struct_info.QueryTagFlags{
				Alias: "u1",
			},
		},
		{
			name: "pk",
			tag:  "pk",
			expected: struct_info.QueryTagFlags{
				IsPK: true,
			},
		},
		{
			name: "omit",
			tag:  "omit",
			expected: struct_info.QueryTagFlags{
				IsOmit: true,
			},
		},
		{
			name: "sort=1",
			tag:  "sort=1",
			expected: struct_info.QueryTagFlags{
				SortPos: 1,
			},
		},
		{
			name: "sort=5",
			tag:  "sort=5",
			expected: struct_info.QueryTagFlags{
				SortPos: 5,
			},
		},
		{
			name: "map single",
			tag:  "map=user_row:u1",
			expected: struct_info.QueryTagFlags{
				RefMap: map[string]string{"user_row": "u1"},
			},
		},
		{
			name: "map multiple",
			tag:  "map=user_row:u1,order_row:o1",
			expected: struct_info.QueryTagFlags{
				RefMap: map[string]string{
					"user_row":  "u1",
					"order_row": "o1",
				},
			},
		},
		{
			name: "combined flags",
			tag:  "from;join=left;alias=u1;pk;sort=2",
			expected: struct_info.QueryTagFlags{
				IsFrom:   true,
				JoinType: "left",
				Alias:    "u1",
				IsPK:     true,
				SortPos:  2,
			},
		},
		{
			name: "full example",
			tag:  "from;join=inner;alias=main;map=ref1:alias1,ref2:alias2;pk;sort=10",
			expected: struct_info.QueryTagFlags{
				IsFrom:   true,
				JoinType: "inner",
				Alias:    "main",
				RefMap: map[string]string{
					"ref1": "alias1",
					"ref2": "alias2",
				},
				IsPK:    true,
				SortPos: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := struct_info.ParseQueryTag(tt.tag)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
