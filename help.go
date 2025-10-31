// Copyright (c) 2025, Mads Moi-Aune <mads@moiaune.dev>
//
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"fmt"
	"os"
)

var loginUsageTmpl = `Usage: %s login --instance INSTANCE_URL --client-id CLIENT_ID
The login command only support OAuth2.0 authentication with public PKCS.

OPTIONS:

  -i, --instance
        Specify the ServiceNow instance name or full URL. snowy will add
        https:// to the value if not present. Must be used in conjuction
        with --client-id.

  --client-id
        Specify the client id for the ServiceNow OAuth2.0 application to
        authenticate with.

        Must be used in conjuction with --instance.
`

var basicUsageTmpl = `snowy is a CLI application for interacting with ServiceNow Table REST API
Usage: %s [COMMAND] [ OPTIONS... ] ARGS
COMMANDS:
  By default snowy will assume that the COMMAND is 'table' if nothing is
  specified.

  table
        Interact with the ServiceNow Table API. This is the default command
        if nothing is specified.

  login
        Start the OAuth2.0 public client PKCS login flow.

  logout
        Removes your authentication profile. Requires you to login next time.

ARGUMENTS:
  ARGS must be either the table_name or a combination of table_name and sys_id
  in the format 'table_name/sys_id'.

OPERATIONS:
  Snowy supports five operations: list, get, insert, update and delete.
  It will try to guess the intended operation by the ARGS and options.
  The only deviation is if you want to delete a record. Then you must
  specify the --delete option (see OPTIONS).

  Examples

    > snowy incident
    This will operate as a list request, to list out records on the incident
    table.

    > snowy incident/af204b6f7560459c8849aa1045b39968
    This will operate a get request to get the specific incident record
    identified by the sys_id.

    > snowy --data '{"short_description": "Hello, World" }' incident
    This will operate as an insert request to the incident table

    > snowy --data '{"impact": 2, "urgency": 2}' incident/af204b6f7560459c8849aa1045b39968
    This will operate as an update request to the specific incident identified
    by the sys_id.

    > snowy --delete incident/af204b6f7560459c8849aa1045b39968
    This will operate as a delete request for the specific incident identified
    by the sys_id. Here we need to explicitly tell snowy to use the "DELETE"
    request method, otherwise it will assume a get operation.

AUTHENTICATION:
  Most calls to ServiceNow Table API requires authentication. Snowy presents
  different ways to authenticate. If credentials are not set by arguments
  (see OPTIONS) then snowy will look for credentials in the following
  environment variables.

  - SNOWY_INSTANCE_URL
  - SNOWY_USERNAME
  - SNOWY_PASSWORD

  If they are not set then it will look for a ~/.snowy or another file if
  specified by --auth-file.

  The snowy credential file must be in the format of:

  <instance_url>
  <username>
  <password>

OPTIONS:
  Options start with one or two dashes. Many of the options require an
  additional value next to them. Some options will only work for certain
  operations. If provided text does not start with a dash, it is presumed to
  be and treated as a table_name or combination of table_name and sys_id
  (see ARGUMENTS).

  SERVICENOW

  -A, --order-asc
        Order the results in ascending order.

  -d, --data
        Data for request body. Can be passed in from stdin.

  --display-value string
        Return field display values (true), actual values (false), or
        both (all) (default "false").

  --exclude-reference-link
        Exclude Table API links for reference fields.

  -f, --fields
        A comma-separated list of fields to return in the response

  --input-display-value
        Set field values using their display value (true) or actual
        value (false) (default: false)

  -l, --limit int
        The maximum number of results returned per page (default 100).

  -o, --order-by string
        A field to order the results by (default sys_created_on).

  --suppress-auto-sys-fields
        True to suppress auto generation of system fields (default: false)

  --suppress-pagination-header
        Supress pagination header.

  --query-no-domain
        True to access data across domains if authorized (default: false)

  -q, --query string
        An encoded query string used to filter the results.

  AUTHENTICATION

  -i, --instance
        Specify the ServiceNow instance name or full URL. snowy will add
        https:// to the value if not present. Must be used in conjuction
        with -u, --user or --client-id.

  -u, --user
        Specify the user name and password to use for API authentication.
        Overrides --auth-file and environment variables. The password will be
        encoded to base64 by snowy. If you only specify the user name, Snowy
        will prompt you for a password.

        Must be used in conjuction with -i, --instance

        Examples:

        > snowy --instance https://dev3843.service-now.com --user username:password incident
        > snowy --instance https://dev3848.service-now.com --user username incident

  --client-id
        Specify the client id for the ServiceNow OAuth2.0 application to
        authenticate with. This will enforce OAuth2 authentication instead of
        Basic Authentication.

        Must be used in conjuction with --instance.

  --auth-file
        By default, Snowy will look for credentials in environment variables
        or ~/.snowy. But you can specify another path to a credential file if
        you want.

        Examples:

        > snowy --auth-file ~/.snowy-test incident
        > snowy --auth-file ~/.snowy-prod incident

  HTTP

  -D, --delete
        Is required when you want to delete a record.

  OTHER

  -h, --help
        Print help
`

func printUsage() {
	fmt.Printf(basicUsageTmpl, os.Args[0])
}

func printLoginUsage() {
	fmt.Printf(loginUsageTmpl, os.Args[0])
}
