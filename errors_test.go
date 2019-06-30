package gqlclient

import (
	"testing"
)

func TestGraphqlErrors(t *testing.T) {
	cases := []struct {
		name      string
		err       GraphqlErrors
		expErrStr string
	}{
		{
			name:      "nil error",
			err:       nil,
			expErrStr: "nil",
		},
		{
			name: "just 1 error",
			err: GraphqlErrors{
				{
					Message: "error-1",
					Locations: []GraphqlErrLoc{
						{
							Line:   1,
							Column: 2,
						},
					},
				},
			},
			expErrStr: "1:2: error-1",
		},
		{
			name: "1 error without a location",
			err: GraphqlErrors{
				{
					Message: "error-1",
				},
			},
			expErrStr: "error-1",
		},
		{
			name: "1 error, 2 locations",
			err: GraphqlErrors{
				{
					Message: "error-1",
					Locations: []GraphqlErrLoc{
						{
							Line:   1,
							Column: 2,
						},
						{
							Line:   10,
							Column: 20,
						},
					},
				},
			},
			expErrStr: "1:2,10:20: error-1",
		},
		{
			name: "2 errors",
			err: GraphqlErrors{
				{
					Message: "error-1",
					Locations: []GraphqlErrLoc{
						{
							Line:   1,
							Column: 2,
						},
					},
				},
				{
					Message: "error-2",
					Locations: []GraphqlErrLoc{
						{
							Line:   1,
							Column: 2,
						},
						{
							Line:   8,
							Column: 9,
						},
					},
				},
			},
			expErrStr: "[0] 1:2: error-1\n[1] 1:2,8:9: error-2\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			errStr := testCase.err.Error()

			if errStr != testCase.expErrStr {
				t.Fatalf("unexpected error string got: %s expected: %s", errStr, testCase.expErrStr)
			}
		})
	}
}
