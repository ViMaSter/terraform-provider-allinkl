package provider

import "strings"

var errorCodeToDescription = map[string]string{
	"kas_access_forbidden": "Invalid `kas_login`/`ALLINKL_KAS_LOGIN` or `kas_auth_data`/`ALLINKL_KAS_AUTH_DATA`",
	"password_syntax_incorrect": "AllInkl requires `dyndns_password` to contain:" + `
- more than 8 characters
- at least 1 lowercase letter
- at least 1 uppercase letter
- at least 1 number
- at least one THSE special characters: /-_#*+!§,()=:.@äöüÄÖÜß
- no simple phrases (123, abc, aaa)`,
	"dyndns_comment_syntax_incorrect": "AllInkl requires `dyndns_comment` to be alphanumeric and between 3 and 64 characters",
	"dyndns_label_not_allowed":        "dyndns_label (subdomains) need to be unique and this label already exists",
}

type Operation string

const (
	OperationCreate Operation = "create"
	OperationRead   Operation = "read"
	OperationUpdate Operation = "update"
)

var operationDescriptions = map[Operation]string{
	OperationCreate: "Creating",
	OperationRead:   "Reading",
	OperationUpdate: "Updating",
}

func reportErrorWithDescription(addError func(summary string, detail string), operation Operation, input string) {
	input = strings.Replace(input, "KasApi: SOAP-ENV:Server: ", "", 1)
	if _, ok := errorCodeToDescription[input]; !ok {
		addError(
			"Error "+operationDescriptions[operation]+" AllInkl DDNS",
			"Unexpected error returned by AllInkl API: "+input+". Please open an issue at https://github.com/ViMaSter/terraform-provider-allinkl/issues",
		)
		return
	}

	addError(
		"Error"+operationDescriptions[operation]+"AllInkl DDNS",
		input+": "+errorCodeToDescription[input],
	)
}
