package allinkl

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

const kasAPIEnvelope = `
<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
    <Body>
        <KasApi xmlns="https://kasserver.com/">
            <Params>%s</Params>
        </KasApi>
    </Body>
</Envelope>`

type KasAPIResponseEnvelope struct {
	XMLName xml.Name   `xml:"Envelope"`
	Body    KasAPIBody `xml:"Body"`
}

type KasAPIBody struct {
	KasAPIResponse *KasResponse `xml:"KasApiResponse"`
	Fault          *Fault       `xml:"Fault"`
}

// ---

type KasRequest struct {
	// Login username
	Login string `json:"kas_login,omitempty"`
	// AuthType `session` or `plain`
	AuthType string `json:"kas_auth_type,omitempty"`
	// AuthData token if AuthType is `session`, password if AuthType is `plain`
	AuthData string `json:"kas_auth_data,omitempty"`
	// Action API function to call
	Action string `json:"kas_action,omitempty"`
	// RequestParams Parameters for the API function
	RequestParams any `json:"KasRequestParams,omitempty"`
}

type DDNSRequest struct {
	// DyndnsComment is a comment for the DDNS entry.
	DyndnsComment string `json:"dyndns_comment,omitempty"`
	// DyndnsPassword is the password for the DDNS entry.
	DyndnsPassword string `json:"dyndns_password,omitempty"`
	// DyndnsZone is the zone/domain for the DDNS entry.
	DyndnsZone string `json:"dyndns_zone"`
	// DyndnsLabel is the label/hostname for the DDNS entry.
	DyndnsLabel string `json:"dyndns_label"`
	// DyndnsTargetIP is the target IP address for the DDNS entry.
	DyndnsTargetIP string `json:"dyndns_target_ip"`
}

type DDNSUpdateRequest struct {
	// DyndnsLogin is the login for the DDNS entry.
	DyndnsLogin string `json:"dyndns_login,omitempty"`
	// DyndnsComment is a comment for the DDNS entry.
	DyndnsComment string `json:"dyndns_comment,omitempty"`
	// DyndnsPassword is the password for the DDNS entry.
	DyndnsPassword string `json:"dyndns_password,omitempty"`
	// DyndnsZone is the zone/domain for the DDNS entry.
	DyndnsZone string `json:"dyndns_zone"`
	// DyndnsLabel is the label/hostname for the DDNS entry.
	DyndnsLabel string `json:"dyndns_label"`
	// DyndnsTargetIP is the target IP address for the DDNS entry.
	DyndnsTargetIP string `json:"dyndns_target_ip"`
}

type GetDDNSUserAPIResponse struct {
	Response GetDDNSUserResponse `json:"Response" mapstructure:"Response"`
}

type GetDDNSUserResponse struct {
	KasFloodDelay float64      `json:"KasFloodDelay" mapstructure:"KasFloodDelay"`
	ReturnInfo    []ReturnInfo `json:"ReturnInfo" mapstructure:"ReturnInfo"`
	ReturnString  string       `json:"ReturnString"`
}

type ReturnInfo struct {
	DyndnsLogin      string `json:"dyndns_login" mapstructure:"dyndns_login"`
	DyndnsComment    string `json:"dyndns_comment" mapstructure:"dyndns_comment"`
	DyndnsLabel      string `json:"dyndns_label" mapstructure:"dyndns_label"`
	DyndnsZone       string `json:"dyndns_zone" mapstructure:"dyndns_zone"`
	DyndnsDualStack  string `json:"dyndns_dual_stack" mapstructure:"dyndns_dual_stack"`
	PWM              string `json:"pwm" mapstructure:"pwm"`
	DyndnsPassword   string `json:"dyndns_password" mapstructure:"dyndns_password"`
	DyndnsTargetIP   string `json:"dyndns_target_ip" mapstructure:"dyndns_target_ip"`
	DyndnsTargetIPv4 string `json:"dyndns_target_ipv4" mapstructure:"dyndns_target_ipv4"`
	DyndnsTargetIPv6 string `json:"dyndns_target_ipv6" mapstructure:"dyndns_target_ipv6"`
	ReturnString     string `json:"ReturnString" mapstructure:"ReturnString"`
}

type AddDDNSUserAPIResponse struct {
	Response AddDDNSUserResponse `json:"Response" mapstructure:"Response"`
}

type AddDDNSUserResponse struct {
	KasFloodDelay float64 `json:"KasFloodDelay" mapstructure:"KasFloodDelay"`
	ReturnInfo    string  `json:"ReturnInfo" mapstructure:"ReturnInfo"`
	ReturnString  string  `json:"ReturnString" mapstructure:"ReturnString"`
}

type DeleteDDNSUserAPIResponse struct {
	Response DeleteDDNSUserResponse `json:"Response" mapstructure:"Response"`
}

type DeleteDDNSUserResponse struct {
	KasFloodDelay float64 `json:"KasFloodDelay" mapstructure:"KasFloodDelay"`
	ReturnInfo    string  `json:"ReturnInfo" mapstructure:"ReturnInfo"`
	ReturnString  string  `json:"ReturnString" mapstructure:"ReturnString"`
}

// helper

// Trimmer trim all XML fields.
type Trimmer struct {
	decoder *xml.Decoder
}

func (tr Trimmer) Token() (xml.Token, error) {
	t, err := tr.decoder.Token()
	if cd, ok := t.(xml.CharData); ok {
		t = xml.CharData(bytes.TrimSpace(cd))
	}
	return t, err
}

// Fault a SOAP fault.
type Fault struct {
	Code    string `xml:"faultcode"`
	Message string `xml:"faultstring"`
	Actor   string `xml:"faultactor"`
}

func (f Fault) Error() string {
	return fmt.Sprintf("%s: %s: %s", f.Actor, f.Code, f.Message)
}

// KasResponse a KAS SOAP response.
type KasResponse struct {
	Return *Item `xml:"return"`
}

// Item an item of the KAS SOAP response.
type Item struct {
	Text  string  `xml:",chardata" json:"text,omitempty"`
	Type  string  `xml:"type,attr" json:"type,omitempty"`
	Raw   string  `xml:"nil,attr" json:"raw,omitempty"`
	Key   *Item   `xml:"key" json:"key,omitempty"`
	Value *Item   `xml:"value" json:"value,omitempty"`
	Items []*Item `xml:"item" json:"item,omitempty"`
}

func decodeXML[T any](reader io.Reader) (*T, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var result T
	err = xml.NewTokenDecoder(Trimmer{decoder: xml.NewDecoder(bytes.NewReader(raw))}).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("decode XML response: %w", err)
	}

	return &result, nil
}
