/*
Copyright Hydrusio Labs Inc. 2016 All Rights Reserved.
Written by mint.zhao.chiu@gmail.com. github.com: https://www.github.com/mintzhao

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package coin

import (
	"errors"
	"time"
	"encoding/base64"
	"github.com/golang/protobuf/proto"
)

// NewTransaction
// if founder is empty, tx is a coinbase transaction
func NewTransaction(founder string) *TX {
	tx := new(TX)
	tx.Version = 1
	tx.Timestamp = time.Now().UTC().Unix()
	tx.Founder = founder
	tx.Txin = make([]*TX_TXIN, 0)
	tx.Txout = make([]*TX_TXOUT, 0)

	return tx
}

// NewTxIn returns a new transaction input
func NewTxIn(owner, prevHash string, prevIdx uint32) *TX_TXIN {
	return &TX_TXIN{
		SourceHash: prevHash,
		Ix:         prevIdx,
		Addr:       owner,
	}
}

// NewTxOut returns a new transaction output
func NewTxOut(value uint64, addr string, until int64) *TX_TXOUT {
	return &TX_TXOUT{
		Value: value,
		Addr:  addr,
		Until: until,
	}
}

// VerifyTx verify tx is valid or not
// If not, error returned
func VerifyTx(tx *TX) error {
	// time check
	if time.Now().UTC().Before(time.Unix(tx.Timestamp, 0).UTC()) {
		return errors.New("tx occur time after now, invalid")
	}

	if tx.Founder == "" {
		return errors.New("no founder transaction")
	}

	if tx.Txout == nil || len(tx.Txout) == 0 {
		return errors.New("transaction output is empty")
	}

	return nil
}


// TxHash generates the Hash for the transaction.
func TxToString(tx *TX) string {
	txBytes, err := proto.Marshal(tx)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(txBytes)
}
