package main

import (
	"encoding/json"

	"github.com/keybase/go-keychain"
)

type Auth struct {
	Password string `json:"password"`
	Secret   string `json:"secret"`
}

func StoreAuth(auth Auth) error {
	data, err := json.Marshal(auth)
	if err != nil {
		return err
	}

	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService("tssh")
	item.SetAccount("tssh")
	item.SetLabel("Teleport Login for tssh")
	item.SetData(data)
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleWhenUnlockedThisDeviceOnly)

	return keychain.AddItem(item)
}

func GetAuth() (*Auth, error) {
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService("tssh")
	query.SetAccount("tssh")
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)
	results, err := keychain.QueryItem(query)
	if err != nil {
		return nil, err
	}

	if len(results) < 1 {
		return nil, nil
	} else {
		auth := &Auth{}

		err := json.Unmarshal(results[0].Data, auth)
		if err != nil {
			return nil, err
		}

		return auth, nil
	}
}

func DeleteAuth() error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService("tssh")
	item.SetAccount("tssh")
	return keychain.DeleteItem(item)
}
