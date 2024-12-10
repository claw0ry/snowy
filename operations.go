package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
)

func doGetOperation(client *Client, opts *CmdOptions) error {
	params := url.Values{}

	if opts.DisplayValue != "false" {
		params.Add("sysparm_display_value", opts.DisplayValue)
	}

	if opts.ExcludeReferenceLink {
		params.Add("sysparm_exclude_reference_link", strconv.FormatBool(opts.ExcludeReferenceLink))
	}

	if opts.Fields != "" {
		params.Add("sysparm_fields", opts.Fields)
	}

	res, err := client.Get(opts.Resource, params)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)
	fmt.Fprintln(os.Stdout, string(data))

	return nil
}

func doListOperation(client *Client, opts *CmdOptions) error {
	params := url.Values{}

	if opts.DisplayValue != "false" {
		params.Add("sysparm_display_value", opts.DisplayValue)
	}

	if opts.ExcludeReferenceLink {
		params.Add("sysparm_exclude_reference_link", strconv.FormatBool(opts.ExcludeReferenceLink))
	}

	if opts.Fields != "" {
		params.Add("sysparm_fields", opts.Fields)
	}

	if opts.SuppressPaginationHeader {
		params.Add("sysparm_suppress_pagination_header", strconv.FormatBool(opts.SuppressPaginationHeader))
	}

	if opts.Limit > 0 {
		params.Add("sysparm_limit", strconv.Itoa(opts.Limit))
	}

	encodedQuery := ""

	if opts.EncodedQuery != "" {
		encodedQuery += opts.EncodedQuery
	}

	orderOperator := "ORDERBYDESC"

	if opts.OrderAsc {
		orderOperator = "ORDERBY"
	}

	if opts.OrderBy != "" {
		encodedQuery += "^" + orderOperator + opts.OrderBy
	}

	if encodedQuery != "" {
		params.Add("sysparm_query", encodedQuery)
	}

	res, err := client.Get(opts.Resource, params)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)
	fmt.Fprintln(os.Stdout, string(data))

	return nil
}

func doInsertOperation(client *Client, opts *CmdOptions) error {
	params := url.Values{}

	if opts.DisplayValue != "false" {
		params.Add("sysparm_display_value", opts.DisplayValue)
	}

	if opts.ExcludeReferenceLink {
		params.Add("sysparm_exclude_reference_link", strconv.FormatBool(opts.ExcludeReferenceLink))
	}

	if opts.InputDisplayValue {
		params.Add("sysparm_input_display_value", strconv.FormatBool(opts.InputDisplayValue))
	}

	if opts.SuppressAutoSysField {
		params.Add("sysparm_suppress_auto_sys_field", strconv.FormatBool(opts.SuppressAutoSysField))
	}

	if opts.Fields != "" {
		params.Add("sysparm_fields", opts.Fields)
	}

	// TODO Support getting raw JSON from stdin if 'opts.Data' is empty
	res, err := client.Post(opts.Resource, params, []byte(opts.Data))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)
	fmt.Fprintln(os.Stdout, string(data))

	return nil
}

func doUpdateOperation(client *Client, opts *CmdOptions) error {
	params := url.Values{}

	if opts.DisplayValue != "false" {
		params.Add("sysparm_display_value", opts.DisplayValue)
	}

	if opts.ExcludeReferenceLink {
		params.Add("sysparm_exclude_reference_link", strconv.FormatBool(opts.ExcludeReferenceLink))
	}

	if opts.InputDisplayValue {
		params.Add("sysparm_input_display_value", strconv.FormatBool(opts.InputDisplayValue))
	}

	if opts.SuppressAutoSysField {
		params.Add("sysparm_suppress_auto_sys_field", strconv.FormatBool(opts.SuppressAutoSysField))
	}

	if opts.Fields != "" {
		params.Add("sysparm_fields", opts.Fields)
	}

	if opts.QueryNoDomain {
		params.Add("sysparm_query_no_domain", strconv.FormatBool(opts.QueryNoDomain))
	}

	// FIXME: Support getting raw JSON from stdin if 'opts.Data' is empty
	res, err := client.Patch(opts.Resource, params, []byte(opts.Data))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)
	fmt.Fprintln(os.Stdout, string(data))

	return nil
}

func doDeleteOperation(client *Client, opts *CmdOptions) error {
	params := url.Values{}

	if opts.QueryNoDomain {
		params.Add("sysparm_query_no_domain", strconv.FormatBool(opts.QueryNoDomain))
	}

	res, err := client.Delete(opts.Resource, params)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)
	fmt.Fprintln(os.Stdout, string(data))

	return nil
}
