package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBoltMetaDB(t *testing.T) {
	t.Run("valid db file", func(t *testing.T) {
		dbFileDir := t.TempDir()
		bMeta, err := NewBoltMetaDB(dbFileDir)

		require.NoError(t, err)
		assert.Equal(t, dbFileDir, bMeta.dbFileDir)
		assert.NotNil(t, bMeta.db)
		assert.NoError(t, bMeta.Close())
	})
}
