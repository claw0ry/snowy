// Copyright (c) 2025, Mads Moi-Aune <mads@moiaune.dev>
//
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

type BasicAuthCredential struct {
	InstanceURL string
	Username    string
	Password    string
}

type basicAuthTransport func(*http.Request) (*http.Response, error)

func (f basicAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newBasicAuthClient(username, password string) *http.Client {
	client := &http.Client{
		Transport: basicAuthTransport(func(req *http.Request) (*http.Response, error) {
			req.SetBasicAuth(username, password)
			req.Header.Add("Accept", "application/json")
			return http.DefaultTransport.RoundTrip(req)
		}),
		Timeout: time.Second * 30,
	}

	return client
}

func credentialsFromFile(c *BasicAuthCredential, fp string) error {
	f, err := os.Open(fp)
	if err != nil {
		return err
	}

	// TODO validate that all fields are set
	s := bufio.NewScanner(f)
	lineN := 0
	for s.Scan() {
		if lineN == 0 {
			c.InstanceURL = s.Text()
		}

		if lineN == 1 {
			c.Username = s.Text()
		}

		if lineN == 2 {
			c.Password = s.Text()
		}

		lineN++
	}

	if lineN < 3 {
		return fmt.Errorf("Not enough info in auth file")
	}

	return nil
}

func loadCredentials(c *BasicAuthCredential, opts *CmdOptions) error {
	if opts.AuthFile != "" {
		if err := credentialsFromFile(c, opts.AuthFile); err != nil {
			return fmt.Errorf("could not load credentials from specified file: %w", err)
		}
		return nil
	}

	instance_url := strings.TrimSpace(opts.Instance)
	if instance_url != "" {
		c.InstanceURL = instance_url
		if !strings.HasPrefix(c.InstanceURL, "https://") {
			c.InstanceURL = "https://" + c.InstanceURL
		}
	}

	user := strings.TrimSpace(opts.User)
	if user != "" {
		if strings.Contains(user, ":") {
			s := strings.Split(user, ":")
			c.Username = s[0]
			c.Password = strings.Join(s[1:], "")
		} else {
			c.Username = user
			fmt.Print("Password: ")
			passwd, err := term.ReadPassword(syscall.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read password from stdin")
			}
			c.Password = strings.TrimSpace(string(passwd))
		}
	}

	return nil
}
