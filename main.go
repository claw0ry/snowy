// Copyright (c) 2025, Mads Moi-Aune <mads@moiaune.dev>
//
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
)

type CmdOptions struct {
	Data                     string
	DisplayValue             string
	EncodedQuery             string
	ExcludeReferenceLink     bool
	Fields                   string
	InputDisplayValue        bool
	Limit                    int
	OrderAsc                 bool
	OrderBy                  string
	QueryNoDomain            bool
	SuppressAutoSysField     bool
	SuppressPaginationHeader bool

	Instance     string
	ClientID     string
	User         string
	AuthFile     string
	ShouldDelete bool

	ShowHelp bool

	Resource string
	Command  Command
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var opts CmdOptions
	if err := parseCommandLine(&opts); err != nil {
		fmt.Printf("ERR: %+v\n", err)
		os.Exit(1)
	}

	if opts.Command == CommandLogin && opts.ShowHelp {
		printLoginUsage()
		os.Exit(0)
	}

	if opts.ShowHelp {
		printUsage()
		os.Exit(0)
	}

	var cmdErr error
	switch opts.Command {
	case CommandLogin:
		cmdErr = doLoginCommand(&opts)
	case CommandLogout:
		cmdErr = doLogoutCommand()
	case CommandTable:
		cmdErr = doTableCommand(&opts)
	default:
		cmdErr = errors.New("unknown command")
	}

	if cmdErr != nil {
		fmt.Fprintf(os.Stderr, "ERR: %+v\n", cmdErr)
		os.Exit(1)
	}

	// 	// stdin, _ := os.Stdin.Stat()
	// 	// if stdin.Mode()&os.ModeNamedPipe != 0 {
	// 	// 	data, _ := io.ReadAll(os.Stdin)
	// 	// 	sysId := strings.TrimSpace(string(data))

	// 	// 	if strings.Contains(sysId, "/") {
	// 	// 		tableCmdOpts.Resource = sysId
	// 	// 	} else {
	// 	// 		tableCmdOpts.Resource = tableCmd.Arg(0) + "/" + sysId
	// 	// 	}
	// 	// } else {
	// 	// 	tableCmdOpts.Resource = tableCmd.Arg(0)
	// 	// }
}

func parseCommandLine(opts *CmdOptions) error {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.StringVar(&opts.Data, "d", "", "")
	f.StringVar(&opts.Data, "data", "", "")
	f.StringVar(&opts.DisplayValue, "display-value", "false", "")
	f.StringVar(&opts.EncodedQuery, "q", "", "")
	f.StringVar(&opts.EncodedQuery, "query", "", "")
	f.BoolVar(&opts.ExcludeReferenceLink, "exclude-reference-link", false, "")
	f.StringVar(&opts.Fields, "f", "", "")
	f.StringVar(&opts.Fields, "fields", "", "")
	f.BoolVar(&opts.InputDisplayValue, "input-display-value", false, "")
	f.IntVar(&opts.Limit, "l", 100, "")
	f.IntVar(&opts.Limit, "limit", 100, "")
	f.BoolVar(&opts.OrderAsc, "A", false, "")
	f.BoolVar(&opts.OrderAsc, "order-asc", false, "")
	f.StringVar(&opts.OrderBy, "o", "", "")
	f.StringVar(&opts.OrderBy, "order-by", "", "")
	f.BoolVar(&opts.QueryNoDomain, "query-no-domain", false, "")
	f.BoolVar(&opts.SuppressAutoSysField, "suppress-auto-sys-fields", false, "")
	f.BoolVar(&opts.SuppressPaginationHeader, "suppress-pagination-header", false, "")
	f.StringVar(&opts.Instance, "i", "", "")
	f.StringVar(&opts.Instance, "instance", "", "")
	f.StringVar(&opts.ClientID, "client-id", "", "")
	f.StringVar(&opts.User, "u", "", "")
	f.StringVar(&opts.User, "user", "", "")
	f.StringVar(&opts.AuthFile, "auth-file", "", "")
	f.BoolVar(&opts.ShouldDelete, "D", false, "")
	f.BoolVar(&opts.ShouldDelete, "delete", false, "")
	f.BoolVar(&opts.ShowHelp, "h", false, "")
	f.BoolVar(&opts.ShowHelp, "help", false, "")

	f.SetOutput(io.Discard)

	var err error
	switch strings.ToLower(os.Args[1]) {
	case "login":
		err = f.Parse(os.Args[2:])
		opts.Command = CommandLogin
	case "logout":
		opts.Command = CommandLogout
	case "table":
		err = f.Parse(os.Args[2:])
		opts.Resource = f.Arg(0)
		opts.Command = CommandTable
	default:
		err = f.Parse(os.Args[1:])
		opts.Resource = f.Arg(0)
		opts.Command = CommandTable
	}

	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil
}

func ensureConfigDir() error {
	dir, err := getConfigDir()
	if err != nil {
		return err
	}

	_, err = os.Stat(dir)
	if err == nil {
		return nil
	}

	if _, ok := err.(*fs.PathError); ok {
		return os.Mkdir(dir, os.FileMode(0700))
	}
	return err
}

func getConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get users config dir")
	}
	return path.Join(dir, "snowy"), nil
}
