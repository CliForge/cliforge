package update

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Version
		wantErr bool
	}{
		{
			name:  "simple version",
			input: "1.2.3",
			want: &Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			wantErr: false,
		},
		{
			name:  "version with v prefix",
			input: "v1.2.3",
			want: &Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			wantErr: false,
		},
		{
			name:  "version with prerelease",
			input: "1.2.3-beta",
			want: &Version{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "beta",
			},
			wantErr: false,
		},
		{
			name:  "version with prerelease and metadata",
			input: "1.2.3-beta+build123",
			want: &Version{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "beta",
				Metadata:   "build123",
			},
			wantErr: false,
		},
		{
			name:  "version with metadata only",
			input: "1.2.3+build123",
			want: &Version{
				Major:    1,
				Minor:    2,
				Patch:    3,
				Metadata: "build123",
			},
			wantErr: false,
		},
		{
			name:    "invalid version - missing patch",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:    "invalid version - non-numeric major",
			input:   "a.2.3",
			wantErr: true,
		},
		{
			name:    "invalid version - empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Major != tt.want.Major || got.Minor != tt.want.Minor || got.Patch != tt.want.Patch ||
				got.Prerelease != tt.want.Prerelease || got.Metadata != tt.want.Metadata {
				t.Errorf("ParseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name    string
		version *Version
		want    string
	}{
		{
			name: "simple version",
			version: &Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			want: "1.2.3",
		},
		{
			name: "version with prerelease",
			version: &Version{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "beta",
			},
			want: "1.2.3-beta",
		},
		{
			name: "version with prerelease and metadata",
			version: &Version{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "beta",
				Metadata:   "build123",
			},
			want: "1.2.3-beta+build123",
		},
		{
			name: "version with metadata only",
			version: &Version{
				Major:    1,
				Minor:    2,
				Patch:    3,
				Metadata: "build123",
			},
			want: "1.2.3+build123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("Version.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{
			name: "equal versions",
			v1:   "1.2.3",
			v2:   "1.2.3",
			want: 0,
		},
		{
			name: "v1 greater major",
			v1:   "2.0.0",
			v2:   "1.9.9",
			want: 1,
		},
		{
			name: "v1 less major",
			v1:   "1.0.0",
			v2:   "2.0.0",
			want: -1,
		},
		{
			name: "v1 greater minor",
			v1:   "1.2.0",
			v2:   "1.1.9",
			want: 1,
		},
		{
			name: "v1 less minor",
			v1:   "1.1.0",
			v2:   "1.2.0",
			want: -1,
		},
		{
			name: "v1 greater patch",
			v1:   "1.2.3",
			v2:   "1.2.2",
			want: 1,
		},
		{
			name: "v1 less patch",
			v1:   "1.2.2",
			v2:   "1.2.3",
			want: -1,
		},
		{
			name: "stable greater than prerelease",
			v1:   "1.2.3",
			v2:   "1.2.3-beta",
			want: 1,
		},
		{
			name: "prerelease less than stable",
			v1:   "1.2.3-beta",
			v2:   "1.2.3",
			want: -1,
		},
		{
			name: "prerelease comparison",
			v1:   "1.2.3-beta",
			v2:   "1.2.3-alpha",
			want: 1,
		},
		{
			name: "metadata ignored",
			v1:   "1.2.3+build1",
			v2:   "1.2.3+build2",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err := ParseVersion(tt.v1)
			if err != nil {
				t.Fatalf("Failed to parse v1: %v", err)
			}
			v2, err := ParseVersion(tt.v2)
			if err != nil {
				t.Fatalf("Failed to parse v2: %v", err)
			}

			if got := v1.Compare(v2); got != tt.want {
				t.Errorf("Version.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionIsNewer(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want bool
	}{
		{
			name: "newer version",
			v1:   "1.2.3",
			v2:   "1.2.2",
			want: true,
		},
		{
			name: "older version",
			v1:   "1.2.2",
			v2:   "1.2.3",
			want: false,
		},
		{
			name: "equal version",
			v1:   "1.2.3",
			v2:   "1.2.3",
			want: false,
		},
		{
			name: "stable newer than prerelease",
			v1:   "1.2.3",
			v2:   "1.2.3-beta",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, _ := ParseVersion(tt.v1)
			v2, _ := ParseVersion(tt.v2)

			if got := v1.IsNewer(v2); got != tt.want {
				t.Errorf("Version.IsNewer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionIsPrerelease(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{
			name:    "stable version",
			version: "1.2.3",
			want:    false,
		},
		{
			name:    "prerelease version",
			version: "1.2.3-beta",
			want:    true,
		},
		{
			name:    "version with metadata only",
			version: "1.2.3+build",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := ParseVersion(tt.version)

			if got := v.IsPrerelease(); got != tt.want {
				t.Errorf("Version.IsPrerelease() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionIsOlder(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want bool
	}{
		{
			name: "older version",
			v1:   "1.2.2",
			v2:   "1.2.3",
			want: true,
		},
		{
			name: "newer version",
			v1:   "1.2.3",
			v2:   "1.2.2",
			want: false,
		},
		{
			name: "equal version",
			v1:   "1.2.3",
			v2:   "1.2.3",
			want: false,
		},
		{
			name: "prerelease older than stable",
			v1:   "1.2.3-beta",
			v2:   "1.2.3",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, _ := ParseVersion(tt.v1)
			v2, _ := ParseVersion(tt.v2)

			if got := v1.IsOlder(v2); got != tt.want {
				t.Errorf("Version.IsOlder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionEqual(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want bool
	}{
		{
			name: "equal versions",
			v1:   "1.2.3",
			v2:   "1.2.3",
			want: true,
		},
		{
			name: "different versions",
			v1:   "1.2.3",
			v2:   "1.2.4",
			want: false,
		},
		{
			name: "equal with prerelease",
			v1:   "1.2.3-beta",
			v2:   "1.2.3-beta",
			want: true,
		},
		{
			name: "equal core but different metadata (should be equal)",
			v1:   "1.2.3+build1",
			v2:   "1.2.3+build2",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, _ := ParseVersion(tt.v1)
			v2, _ := ParseVersion(tt.v2)

			if got := v1.Equal(v2); got != tt.want {
				t.Errorf("Version.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionIsStable(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{
			name:    "stable version",
			version: "1.2.3",
			want:    true,
		},
		{
			name:    "prerelease version",
			version: "1.2.3-beta",
			want:    false,
		},
		{
			name:    "version with metadata only",
			version: "1.2.3+build",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := ParseVersion(tt.version)

			if got := v.IsStable(); got != tt.want {
				t.Errorf("Version.IsStable() = %v, want %v", got, tt.want)
			}
		})
	}
}
