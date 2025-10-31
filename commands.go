// Copyright (c) 2025, Mads Moi-Aune <mads@moiaune.dev>
//
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"golang.org/x/oauth2"
)

type Command int

const (
	CommandLogin Command = iota
	CommandLogout
	CommandTable
	CommandUnknown
)

func doLoginCommand(opts *CmdOptions) error {
	instance_url := strings.TrimSpace(opts.Instance)
	client_id := strings.TrimSpace(opts.ClientID)

	if instance_url == "" {
		return fmt.Errorf("Flag '--instance' is either not defined or empty.")
	}

	if client_id == "" {
		return fmt.Errorf("Flag '--client-id' is either not defined or empty.")
	}

	redirectUrl := "http://localhost:1914"
	oauthCfg := &oauth2.Config{
		ClientID:     client_id,
		ClientSecret: "", // empty for PKCE public client
		Endpoint: oauth2.Endpoint{
			AuthURL:  instance_url + "/oauth_auth.do",
			TokenURL: instance_url + "/oauth_token.do",
		},
		RedirectURL: redirectUrl,
		Scopes:      nil,
	}

	token, err := authorizeInteractive(context.Background(), oauthCfg)
	if err != nil {
		return err
	}

	profile := OAuth2Profile{
		InstanceURL: instance_url,
		Token:       token,
		Config:      oauthCfg,
	}
	profile.Save()

	return nil
}

func doLogoutCommand() error {
	configDir, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config dir: %+v", err)
	}
	err = os.Remove(path.Join(configDir, "profile.json"))
	if err != nil {
		return fmt.Errorf("failed to delete profile.json: %+v", err)
	}
	return nil
}

func doTableCommand(opts *CmdOptions) error {
	var creds BasicAuthCredential
	if err := loadCredentials(&creds, opts); err != nil {
		return err
	}

	var client *Client
	var err error
	switch getAuthMethod(opts) {
	case AuthBasic:
		client = NewBasicAuthClient(&creds)
	case AuthOAuth2:
		client, err = NewOAuth2Client(context.Background())
	default:
		err = fmt.Errorf("unknown auth method")
	}
	if err != nil {
		return err
	}

	switch presumeOperation(opts) {
	case OperationGet:
		return doGetOperation(client, opts)
	case OperationList:
		return doListOperation(client, opts)
	case OperationInsert:
		return doInsertOperation(client, opts)
	case OperationUpdate:
		return doUpdateOperation(client, opts)
	case OperationDelete:
		return doDeleteOperation(client, opts)
	}

	return fmt.Errorf("unknown operation")
}

func getAuthMethod(opts *CmdOptions) AuthMode {
	if opts.AuthFile != "" || opts.User != "" {
		return AuthBasic
	}

	return AuthOAuth2
}
