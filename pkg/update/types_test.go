package update

import (
	"testing"
)

func TestUpdateStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status UpdateStatus
		want   string
	}{
		{
			name:   "unknown",
			status: UpdateStatusUnknown,
			want:   "unknown",
		},
		{
			name:   "up to date",
			status: UpdateStatusUpToDate,
			want:   "up-to-date",
		},
		{
			name:   "available",
			status: UpdateStatusAvailable,
			want:   "available",
		},
		{
			name:   "downloading",
			status: UpdateStatusDownloading,
			want:   "downloading",
		},
		{
			name:   "installing",
			status: UpdateStatusInstalling,
			want:   "installing",
		},
		{
			name:   "failed",
			status: UpdateStatusFailed,
			want:   "failed",
		},
		{
			name:   "invalid status",
			status: UpdateStatus(999),
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("UpdateStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
