package update

import (
	"testing"
)

func TestUpdateStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   string
	}{
		{
			name:   "unknown",
			status: StatusUnknown,
			want:   "unknown",
		},
		{
			name:   "up to date",
			status: StatusUpToDate,
			want:   "up-to-date",
		},
		{
			name:   "available",
			status: StatusAvailable,
			want:   "available",
		},
		{
			name:   "downloading",
			status: StatusDownloading,
			want:   "downloading",
		},
		{
			name:   "installing",
			status: StatusInstalling,
			want:   "installing",
		},
		{
			name:   "failed",
			status: StatusFailed,
			want:   "failed",
		},
		{
			name:   "invalid status",
			status: Status(999),
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("Status.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
