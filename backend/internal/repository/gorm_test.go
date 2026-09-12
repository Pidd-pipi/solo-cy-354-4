package repository

import (
	"errors"
	"testing"

	"github.com/lp/campus-market/internal/util"
	"gorm.io/gorm"
)

var boomErr = errors.New("boom")

func TestNormalizeError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "nil", err: nil, want: nil},
		{name: "record not found", err: gorm.ErrRecordNotFound, want: util.ErrNotFound},
		{name: "other preserved", err: boomErr, want: boomErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeError(tt.err)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if !errors.Is(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
