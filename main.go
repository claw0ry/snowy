package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"strings"
	"syscall"

	"golang.org/x/term"
)

var basicUsageTmpl = `snowy is a CLI application for interacting with ServiceNow Table REST API
Usage: %s [ OPTIONS... ] ARGS
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
        with -u, --user

  -u, --user
        Specify the user name and password to use for API authentication.
        Overrides --auth-file and environment variables. The password will be
        encoded to base64 by snowy. If you only specify the user name, Snowy
        will prompt you for a password.

        Must be used in conjuction with -i, --instance

        Examples:

        > snowy --instance https://dev3843.service-now.com --user username:password incident
        > snowy --instance https://dev3848.service-now.com --user username incident

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
	User         string
	AuthFile     string
	ShouldDelete bool

	ShowHelp bool

	Resource string
}

type Operation string

var (
	OperationGet    Operation = "get"
	OperationList   Operation = "list"
	OperationInsert Operation = "insert"
	OperationUpdate Operation = "update"
	OperationDelete Operation = "delete"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var opts CmdOptions
	if err := initializeFlags(&opts); err != nil {
		fmt.Printf("ERR: %+v\n", err)
		os.Exit(1)
	}

	if opts.ShowHelp {
		printUsage()
		os.Exit(0)
	}

	var credentials Credentials
	if err := loadCredentials(&credentials, &opts); err != nil {
		fmt.Printf("ERR: %+v\n", err)
		os.Exit(1)
	}

	client := NewClient(&credentials)

	var cmdErr error
	switch presumeOperation(&opts) {
	case OperationGet:
		cmdErr = doGetOperation(&client, &opts)
	case OperationList:
		cmdErr = doListOperation(&client, &opts)
	case OperationInsert:
		cmdErr = doInsertOperation(&client, &opts)
	case OperationUpdate:
		cmdErr = doUpdateOperation(&client, &opts)
	case OperationDelete:
		cmdErr = doDeleteOperation(&client, &opts)
	}

	if cmdErr != nil {
		fmt.Printf("ERR: %+v\n", cmdErr)
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

func initializeFlags(opts *CmdOptions) error {
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
	f.StringVar(&opts.User, "u", "", "")
	f.StringVar(&opts.User, "user", "", "")
	f.StringVar(&opts.AuthFile, "auth-file", "", "")
	f.BoolVar(&opts.ShouldDelete, "D", false, "")
	f.BoolVar(&opts.ShouldDelete, "delete", false, "")
	f.BoolVar(&opts.ShowHelp, "h", false, "")
	f.BoolVar(&opts.ShowHelp, "help", false, "")

	f.SetOutput(io.Discard)

	err := f.Parse(os.Args[1:])
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	opts.Resource = f.Arg(0)
	return nil
}

// TODO Maybe return some error message if args are not matching with the operation?
// This is really only for --delete. It should display an error if --delete is set when a resource only
// contains a table_name
func presumeOperation(opts *CmdOptions) Operation {
	if strings.Contains(opts.Resource, "/") {
		if opts.ShouldDelete {
			return OperationDelete
		}

		if opts.Data != "" {
			return OperationUpdate
		}
		return OperationGet
	}

	if opts.Data != "" {
		return OperationInsert
	}
	return OperationList
}

func credentialsFromEnv(c *Credentials) error {
	instanceURL := os.Getenv("SNOWY_INSTANCE_URL")
	if instanceURL == "" {
		return fmt.Errorf("Environment variable 'SNOWY_INSTANCE_URL' is not set")
	}

	username := os.Getenv("SNOWY_USERNAME")
	if username == "" {
		return fmt.Errorf("Environment variable 'SNOWY_USERNAME' is not set")
	}

	password := os.Getenv("SNOWY_PASSWORD")
	if password == "" {
		return fmt.Errorf("Environment variable 'SNOWY_PASSWORD' is not set")
	}

	c.InstanceURL = instanceURL
	c.Username = username
	c.Password = password

	return nil
}

func credentialsFromFile(c *Credentials, fp string) error {
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

func loadCredentials(c *Credentials, opts *CmdOptions) error {
	if opts.Instance != "" && opts.User != "" {
		if !strings.HasPrefix(opts.Instance, "https://") {
			opts.Instance = "https://" + opts.Instance
		}

		c.InstanceURL = opts.Instance
		if strings.Contains(opts.User, ":") {
			s := strings.Split(opts.User, ":")
			c.Username = s[0]
			c.Password = strings.Join(s[1:], "")
		} else {
			c.Username = opts.User
			fmt.Print("Password: ")
			passwd, err := term.ReadPassword(syscall.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read password from stdin")
			}

			c.Password = strings.TrimSpace(string(passwd))
		}
		return nil
	}

	_ = credentialsFromEnv(c)

	if c.InstanceURL == "" {
		if opts.AuthFile != "" {
			err := credentialsFromFile(c, opts.AuthFile)
			if err != nil {
				return fmt.Errorf("could not load credentials from specified file: %w", err)
			}
		} else {
			usr, err := user.Current()
			if err != nil {
				return fmt.Errorf("failed to get current user")
			}

			authfile := path.Join(usr.HomeDir, ".snowy")
			err = credentialsFromFile(c, authfile)
			if err != nil {
				return fmt.Errorf("could not load credentials from file: %w", err)
			}
		}
	}

	return nil
}

func printUsage() {
	fmt.Printf(basicUsageTmpl, os.Args[0])
}
