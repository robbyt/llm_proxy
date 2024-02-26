package writers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUniqueFileName(t *testing.T) {
	dir := t.TempDir()
	fileName := createUniqueFileName(dir, "test", "txt", 0)
	assert.Equal(t, fileName, dir+"/test.txt")

	// Create a file to test the unique filename generation
	_, err := os.Create(fileName)
	require.NoError(t, err)

	fileName2 := createUniqueFileName(dir, "test", "txt", 0)
	assert.Equal(t, fileName2, dir+"/test-1.txt")
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()

	assert.False(t, fileExists(dir+"/nonexistent.txt"))

	_, err := os.Create(dir + "/existent.txt")
	require.NoError(t, err)

	assert.True(t, fileExists(dir+"/existent.txt"))
}

func TestCreateNewFileFromFilename(t *testing.T) {
	dir := t.TempDir()

	file, err := createNewFileFromFilename(dir + "/test.txt")
	require.NoError(t, err)
	defer file.Close()

	_, err = os.Stat(dir + "/test.txt")
	assert.NoError(t, err)
}

func TestDirExistsOrCreate(t *testing.T) {
	dir := t.TempDir()

	err := dirExistsOrCreate(dir + "/subdir")
	assert.NoError(t, err)

	_, err = os.Stat(dir + "/subdir")
	assert.NoError(t, err)
}
