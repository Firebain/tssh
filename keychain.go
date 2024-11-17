package main

import (
	"strings"

	"github.com/keybase/go-keychain"
)

type Auth struct {
	password string
	secret   string
}

func StoreAuth(auth Auth) error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService("tssh")
	item.SetAccount("tssh")
	item.SetLabel("Teleport Login for tssh")
	item.SetData([]byte(auth.password + "\n" + auth.secret))
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
		auth := strings.Split(string(results[0].Data), "\n")

		return &Auth{
			password: auth[0],
			secret:   auth[1],
		}, nil
	}
}

func DeleteAuth() error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService("tssh")
	item.SetAccount("tssh")
	return keychain.DeleteItem(item)
}