package utils

import (
	"reflect"
	"testing"
)

func TestPingDelaySet_FilterLossRate(t *testing.T) {
	tests := []struct {
		name     string
		s        PingDelaySet
		wantData PingDelaySet
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotData := tt.s.FilterLossRate(); !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("FilterLossRate() = %v, want %v", gotData, tt.wantData)
			}
		})
	}
}
