package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBadgerDB(t *testing.T) {
	tempDir := t.TempDir()

	badgerDB, err := NewBadgerDB("test", tempDir+"/test.db")
	assert.NoError(t, err)
	assert.NotNil(t, badgerDB)
	assert.Equal(t, "test", badgerDB.identifier)
	assert.Equal(t, tempDir+"/test.db", badgerDB.DBFileName)
}

func TestBadgerDB_BadFileName(t *testing.T) {
	_, err := NewBadgerDB("test", "")
	require.Error(t, err)
}

func TestBadgerDB_SetAndGetBytes(t *testing.T) {
	tempDir := t.TempDir()

	badgerDB, _ := NewBadgerDB("test", tempDir+"/test.db")
	defer badgerDB.Close()

	tests := []struct {
		name string
		key  []byte
		val  []byte
		want []byte
	}{
		{
			name: "Test with typical key/val",
			key:  []byte("key"),
			val:  []byte("val"),
		},
		{
			name: "Test with empty key",
			key:  []byte(""),
			val:  []byte("val"),
		},
		{
			name: "Test with a space key",
			key:  []byte(" "),
			val:  []byte("val2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// delete before starting, should be idempotent
			err := badgerDB.Delete(tt.key)
			require.NoError(t, err)

			value, err := badgerDB.GetBytesSafe(tt.key)
			require.NoError(t, err)
			assert.Nil(t, value)

			err = badgerDB.SetBytes(tt.key, tt.val)
			require.NoError(t, err)

			value, err = badgerDB.GetBytes(tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.val, value)

			value, err = badgerDB.GetStr(string(tt.key))
			require.NoError(t, err)
			assert.Equal(t, tt.val, value)

			// delete again, should be idempotent
			err = badgerDB.Delete(tt.key)
			require.NoError(t, err)

			err = badgerDB.SetStr(string(tt.key), string(tt.val))
			require.NoError(t, err)

			value, err = badgerDB.GetStrSafe(string(tt.key))
			require.NoError(t, err)
			assert.Equal(t, tt.val, value)
		})
	}
}

func TestBadgerDB_Close(t *testing.T) {
	tempDir := t.TempDir()

	badgerDB, _ := NewBadgerDB("test", tempDir+"/test.db")
	err := badgerDB.Close()
	assert.NoError(t, err)

	// close again, should be idempotent
	err = badgerDB.Close()
	assert.NoError(t, err)
}

func TestBadgerDB_CloseBusted(t *testing.T) {
	tempDir := t.TempDir()

	badgerDB, _ := NewBadgerDB("test", tempDir+"/test.db")
	os.RemoveAll(tempDir)

	// close a file that has been deleted should return an error
	err := badgerDB.Close()
	assert.Error(t, err)
}
