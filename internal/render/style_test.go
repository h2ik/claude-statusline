package render

import "testing"

func TestLineData_IsEmpty(t *testing.T) {
	tests := []struct {
		name      string
		line      LineData
		wantLeft  bool
		wantRight bool
	}{
		{
			name:      "empty line",
			line:      LineData{},
			wantLeft:  true,
			wantRight: true,
		},
		{
			name:      "left only",
			line:      LineData{Left: []string{"a"}},
			wantLeft:  false,
			wantRight: true,
		},
		{
			name:      "right only",
			line:      LineData{Right: []string{"b"}},
			wantLeft:  true,
			wantRight: false,
		},
		{
			name:      "both sides",
			line:      LineData{Left: []string{"a"}, Right: []string{"b"}},
			wantLeft:  false,
			wantRight: false,
		},
		{
			name:      "with names",
			line:      LineData{Left: []string{"a"}, LeftNames: []string{"repo_info"}, Right: []string{"b"}, RightNames: []string{"time_display"}},
			wantLeft:  false,
			wantRight: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(tt.line.Left) == 0; got != tt.wantLeft {
				t.Errorf("Left empty = %v, want %v", got, tt.wantLeft)
			}
			if got := len(tt.line.Right) == 0; got != tt.wantRight {
				t.Errorf("Right empty = %v, want %v", got, tt.wantRight)
			}
		})
	}
}
