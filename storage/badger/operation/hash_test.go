// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package operation

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/dapperlabs/flow-go/crypto"
	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashInsertRetrieve(t *testing.T) {

	dir := filepath.Join(os.TempDir(), fmt.Sprintf("flow-test-db-%d", rand.Uint64()))
	db, err := badger.Open(badger.DefaultOptions(dir).WithLogger(nil))
	require.Nil(t, err)

	number := uint64(1337)
	expected := crypto.Hash{0x01, 0x02, 0x03}

	err = db.Update(InsertHash(number, expected))
	require.Nil(t, err)

	var actual crypto.Hash
	err = db.View(RetrieveHash(number, &actual))
	require.Nil(t, err)

	assert.Equal(t, expected, actual)
}
