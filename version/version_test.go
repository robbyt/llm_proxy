package version

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	match, err := regexp.MatchString("[^\\d+\\.]", Version)
	require.NoError(t, err, "Error matching Version regex")
	require.False(t, match, "Version should contain only the main version")

	match, err = regexp.MatchString("[^a-z\\d]", Prerelease)
	require.NoError(t, err, "Error matching Prerelease regex")
	require.False(t, match, "Prerelease should contain only letters and numbers")

	require.Empty(t, SemVer.Prerelease(), "SemVer should not include prerelease information")

	require.Contains(t, String(), Prerelease, "Full version string should include prerelease information")
}
