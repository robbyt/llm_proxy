package fileUtils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUniqueFileName(t *testing.T) {
	dir := t.TempDir()
	fileName := CreateUniqueFileName(dir, "test", "txt", 0)
	assert.Equal(t, fileName, dir+"/test.txt")

	// Create a file to test the unique filename generation
	_, err := os.Create(fileName)
	require.NoError(t, err)

	fileName2 := CreateUniqueFileName(dir, "test", "txt", 0)
	assert.Equal(t, fileName2, dir+"/test-1.txt")
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()

	assert.False(t, FileExists(dir+"/nonexistent.txt"))

	_, err := os.Create(dir + "/existent.txt")
	require.NoError(t, err)

	assert.True(t, FileExists(dir+"/existent.txt"))
}

func TestCreateNewFileFromFilename(t *testing.T) {
	dir := t.TempDir()

	file, err := CreateNewFileFromFilename(dir + "/test.txt")
	require.NoError(t, err)
	defer file.Close()

	_, err = os.Stat(dir + "/test.txt")
	assert.NoError(t, err)
}

func TestDirExistsOrCreate(t *testing.T) {
	dir := t.TempDir()

	err := DirExistsOrCreate(dir + "/subdir")
	assert.NoError(t, err)

	_, err = os.Stat(dir + "/subdir")
	assert.NoError(t, err)
}

func TestRelocateExistingFileIfExists(t *testing.T) {
	// Create a temporary directory for testing
	dir := t.TempDir()

	// Create a temporary file
	file, err := os.CreateTemp(dir, "test")
	assert.NoError(t, err)
	fileName := file.Name()

	// Close and remove the file
	assert.NoError(t, file.Close())
	assert.NoError(t, os.Remove(fileName))
	assert.False(t, FileExists(fileName))

	// Test with a file that does not exist
	err = RelocateExistingFileIfExists(fileName)
	assert.NoError(t, err)

	// Create the file again
	file, err = os.Create(fileName)
	assert.NoError(t, err)
	file.Close()

	// Test with a file that does exist
	err = RelocateExistingFileIfExists(fileName)
	assert.NoError(t, err)

	// Check that the original file does not exist
	assert.False(t, FileExists(fileName))

	// Check that a new file exists
	files, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
}

func TestConvertURLtoFileName(t *testing.T) {
	tests := []struct {
		name      string
		dbFileDir string
		url       string
		want      string
	}{
		{
			name:      "Test with https URL",
			dbFileDir: "/tmp",
			url:       "https://example.com/test?param=value",
			want:      "/tmp/ZXhhbXBsZS5jb20vdGVzdD9wYXJhbT12YWx1ZQ==",
		},
		{
			name:      "Test with http URL",
			dbFileDir: "/tmp",
			url:       "http://example.com/test?param=value&param2=value2",
			want:      "/tmp/ZXhhbXBsZS5jb20vdGVzdD9wYXJhbT12YWx1ZSZwYXJhbTI9dmFsdWUy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertIDtoFileName(tt.dbFileDir, tt.url)
			assert.Equal(t, tt.want, got)
		})
	}
}
