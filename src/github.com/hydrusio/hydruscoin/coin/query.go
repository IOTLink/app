/*
Copyright Mojing Inc. 2016 All Rights Reserved.
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
	"github.com/golang/protobuf/proto"
)

func (coin *Hydruscoin) queryAddrs(store Store, args []string) ([]byte, error) {
	if len(args) < 2 || args[1] == "" {
		return nil, ErrInvalidArgs
	}

	results := &QueryAddrResults{
		Accounts: make(map[string]*Account),
	}

	args = args[1:]
	for _, addr := range args {
		account, err := store.GetAccount(addr)
		if err != nil {
			logger.Warningf("store.GetAccount() return error: %v", err)
			continue
		}

		results.Accounts[addr] = account
		logger.Debugf("query addr[%s] account: %+v", addr, account)
	}

	return proto.Marshal(results)
}

func (coin *Hydruscoin) queryTx(store Store, args []string) ([]byte, error) {
	if len(args) != 2 || args[1] == "" {
		return nil, ErrInvalidArgs
	}

	tx, _, err := store.GetTx(args[1])
	if err != nil {
		logger.Errorf("get tx info error: %v", err)
		return nil, err
	}
	logger.Debugf("query tx: %+v", tx)

	return proto.Marshal(tx)
}

func (coin *Hydruscoin) queryCoin(store Store, args []string) ([]byte, error) {
	if len(args) != 2 ||  args[1] != "" {
		return nil, ErrInvalidArgs
	}

	coinInfo, err := store.GetCoinInfo()
	if err != nil {
		logger.Errorf("Error get coin info: %v", err)
		return nil, err
	}

	logger.Debugf("query lepuscoin info: %+v", coinInfo)
	return proto.Marshal(coinInfo)
}
