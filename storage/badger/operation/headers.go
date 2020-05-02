// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package operation

import (
	"github.com/dgraph-io/badger/v2"

	"github.com/dapperlabs/flow-go/model/flow"
)

func InsertHeader(headerID flow.Identifier, header *flow.Header) func(*badger.Txn) error {
	return insert(makePrefix(codeHeader, headerID), header)
}

func CheckHeader(blockID flow.Identifier, exists *bool) func(*badger.Txn) error {
	return check(makePrefix(codeHeader, blockID), exists)
}

func RetrieveHeader(blockID flow.Identifier, header *flow.Header) func(*badger.Txn) error {
	return retrieve(makePrefix(codeHeader, blockID), header)
}

func IndexHeaderByCollection(collectionID, headerID flow.Identifier) func(*badger.Txn) error {
	return insert(makePrefix(codeIndexHeaderByCollection, collectionID), headerID)
}

func LookupBlockIDByCollectionID(collectionID flow.Identifier, headerID *flow.Identifier) func(*badger.Txn) error {
	return retrieve(makePrefix(codeIndexHeaderByCollection, collectionID), headerID)
}

func IndexBlockHeight(height uint64, blockID flow.Identifier) func(*badger.Txn) error {
	return insert(makePrefix(codeHeightToBlock, height), blockID)
}

func LookupBlockHeight(height uint64, blockID *flow.Identifier) func(*badger.Txn) error {
	return retrieve(makePrefix(codeHeightToBlock, height), blockID)
}
