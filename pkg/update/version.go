// Package update provides self-update capabilities for CliForge-generated CLIs.
package update

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Metadata   string
}

// ParseVersion parses a semantic version string (e.g., "1.2.3", "1.2.3-beta", "1.2.3+build").
func ParseVersion(v string) (*Version, error) {
	if v == "" {
		return nil, fmt.Errorf("version string cannot be empty")
	}

	// Remove 'v' prefix if present
	v = strings.TrimPrefix(v, "v")

	// Split metadata (+build)
	parts := strings.SplitN(v, "+", 2)
	v = parts[0]
	metadata := ""
	if len(parts) == 2 {
		metadata = parts[1]
	}

	// Split prerelease (-beta, -rc.1, etc.)
	parts = strings.SplitN(v, "-", 2)
	v = parts[0]
	prerelease := ""
	if len(parts) == 2 {
		prerelease = parts[1]
	}

	// Parse major.minor.patch
	versionParts := strings.Split(v, ".")
	if len(versionParts) != 3 {
		return nil, fmt.Errorf("invalid version format: expected major.minor.patch, got %s", v)
	}

	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err := strconv.Atoi(versionParts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %w", err)
	}

	return &Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: prerelease,
		Metadata:   metadata,
	}, nil
}

// String returns the string representation of the version.
func (v *Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		s += "-" + v.Prerelease
	}
	if v.Metadata != "" {
		s += "+" + v.Metadata
	}
	return s
}

// Compare compares this version with another.
// Returns:
//   -1 if v < other
//    0 if v == other
//    1 if v > other
func (v *Version) Compare(other *Version) int {
	// Compare major
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	// Compare minor
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	// Compare patch
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// Compare prerelease
	// According to semver: 1.0.0-alpha < 1.0.0
	// A version with prerelease is less than one without
	if v.Prerelease == "" && other.Prerelease != "" {
		return 1
	}
	if v.Prerelease != "" && other.Prerelease == "" {
		return -1
	}

	// Both have prerelease, compare lexicographically
	if v.Prerelease != other.Prerelease {
		if v.Prerelease < other.Prerelease {
			return -1
		}
		return 1
	}

	// Metadata is ignored in version precedence
	return 0
}

// IsNewer returns true if this version is newer than the other.
func (v *Version) IsNewer(other *Version) bool {
	return v.Compare(other) > 0
}

// IsOlder returns true if this version is older than the other.
func (v *Version) IsOlder(other *Version) bool {
	return v.Compare(other) < 0
}

// Equal returns true if this version equals the other.
func (v *Version) Equal(other *Version) bool {
	return v.Compare(other) == 0
}

// IsPrerelease returns true if this is a prerelease version.
func (v *Version) IsPrerelease() bool {
	return v.Prerelease != ""
}

// IsStable returns true if this is a stable version (no prerelease).
func (v *Version) IsStable() bool {
	return !v.IsPrerelease()
}
