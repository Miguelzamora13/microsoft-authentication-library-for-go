// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package main

import (
	"context"
	"fmt"
	"reflect"
	"time"

	msalgo "github.com/AzureAD/microsoft-authentication-library-for-go/src/msal"
	log "github.com/sirupsen/logrus"
)

func deviceCodeCallback(deviceCodeResult msalgo.IDeviceCodeResult) {
	log.Infof(deviceCodeResult.GetMessage())
}

func setCancelTimeout(seconds int, cancelChannel chan bool) {
	time.Sleep(time.Duration(seconds) * time.Second)
	cancelChannel <- true
}

func tryDeviceCodeFlow() {
	cancelTimeout := 100 //Change this for cancel timeout
	cancelCtx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(cancelTimeout)*time.Second)
	defer cancelFunc()
	deviceCodeParams := msalgo.CreateAcquireTokenDeviceCodeParameters(cancelCtx, config.Scopes, deviceCodeCallback)
	resultChannel := make(chan msalgo.IAuthenticationResult)
	errChannel := make(chan error)
	go func() {
		result, err := publicClientApp.AcquireTokenByDeviceCode(deviceCodeParams)
		errChannel <- err
		resultChannel <- result
	}()
	err = <-errChannel
	if err != nil {
		log.Fatal(err)
	}
	result := <-resultChannel
	fmt.Println("Access token is " + result.GetAccessToken())
}

func acquireTokenDeviceCode() {
	config := createConfig("config.json")
	pcaParams := createPCAParams(config.ClientID, config.Authority)
	publicClientApp, err := msalgo.CreatePublicClientApplication(pcaParams)
	if err != nil {
		log.Fatal(err)
	}
	publicClientApp.SetCacheAccessor(cacheAccessor)
	var userAccount msalgo.IAccount
	accounts := publicClientApp.GetAccounts()
	for _, account := range accounts {
		if account.GetUsername() == config.Username {
			userAccount = account
		}
	}
	if reflect.ValueOf(userAccount).IsNil() {
		log.Info("No valid account found")
		tryDeviceCodeFlow()
	} else {
		silentParams := msalgo.CreateAcquireTokenSilentParameters(config.Scopes, userAccount)
		result, err := publicClientApp.AcquireTokenSilent(silentParams)
		if err != nil {
			log.Info(err)
			tryDeviceCodeFlow()
		} else {
			fmt.Println("Access token is " + result.GetAccessToken())
		}
	}
}
