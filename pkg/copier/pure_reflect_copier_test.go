package copier

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func ToPtr[T any](t T) *T {
	return &t
}

func TestReflectCopier_CopyTo(t *testing.T) {
	testCases := []struct {
		name     string
		copyFunc func() (any, error)
		wantDst  any
		wantErr  error
	}{
		{
			name: "simple struct",
			copyFunc: func() (any, error) {
				dst := &SimpleDst{}
				err := CopyTo(&SimpleSrc{
					Name:    "tracydzf",
					Age:     ToPtr[int](18),
					Friends: []string{"Tom", "Jerry"},
				}, dst)
				return dst, err
			},
			wantDst: &SimpleDst{
				Name:    "tracydzf",
				Age:     ToPtr[int](18),
				Friends: []string{"Tom", "Jerry"},
			},
		}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.copyFunc()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantDst, res)
		})
	}
}

type SimpleSrc struct {
	Name    string
	Age     *int
	Friends []string
}

type SimpleDst struct {
	Name    string
	Age     *int
	Friends []string
}
