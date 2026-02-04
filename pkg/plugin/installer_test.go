package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSource_Good_OrgRepo(t *testing.T) {
	org, repo, version, err := ParseSource("host-uk/core-plugin")
	assert.NoError(t, err)
	assert.Equal(t, "host-uk", org)
	assert.Equal(t, "core-plugin", repo)
	assert.Equal(t, "", version)
}

func TestParseSource_Good_OrgRepoVersion(t *testing.T) {
	org, repo, version, err := ParseSource("host-uk/core-plugin@v1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, "host-uk", org)
	assert.Equal(t, "core-plugin", repo)
	assert.Equal(t, "v1.0.0", version)
}

func TestParseSource_Good_VersionWithoutPrefix(t *testing.T) {
	org, repo, version, err := ParseSource("org/repo@1.2.3")
	assert.NoError(t, err)
	assert.Equal(t, "org", org)
	assert.Equal(t, "repo", repo)
	assert.Equal(t, "1.2.3", version)
}

func TestParseSource_Bad_Empty(t *testing.T) {
	_, _, _, err := ParseSource("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source is empty")
}

func TestParseSource_Bad_NoSlash(t *testing.T) {
	_, _, _, err := ParseSource("just-a-name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestParseSource_Bad_TooManySlashes(t *testing.T) {
	_, _, _, err := ParseSource("a/b/c")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestParseSource_Bad_EmptyOrg(t *testing.T) {
	_, _, _, err := ParseSource("/repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestParseSource_Bad_EmptyRepo(t *testing.T) {
	_, _, _, err := ParseSource("org/")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestParseSource_Bad_EmptyVersion(t *testing.T) {
	_, _, _, err := ParseSource("org/repo@")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version is empty")
}
