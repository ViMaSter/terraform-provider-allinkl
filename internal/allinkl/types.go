package allinkl

import (
	"bytes"
	"encoding/json"
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

// Records API (scaffolded to map to DDNS for now)
type DNSRequest struct {
	ZoneHost   string `json:"zone_host"`
	RecordType string `json:"record_type"`
	RecordName string `json:"record_name"`
	RecordData string `json:"record_data"`
	RecordAux  int64  `json:"record_aux"`
}

type DNSUpdateRequest struct {
	ZoneHost   string `json:"zone_host"`
	RecordId   int64  `json:"record_id"`
	RecordType string `json:"record_type"`
	RecordName string `json:"record_name"`
	RecordData string `json:"record_data"`
	RecordAux  int64  `json:"record_aux"`
}

type GetDNSAPIResponse struct {
	Response GetDNSResponse `json:"Response" mapstructure:"Response"`
}

type GetDNSResponse struct {
	KasFloodDelay float64            `json:"KasFloodDelay" mapstructure:"KasFloodDelay"`
	ReturnInfo    []GetDNSReturnInfo `json:"ReturnInfo" mapstructure:"ReturnInfo"`
	ReturnString  string             `json:"ReturnString"`
}

type GetDNSReturnInfo struct {
	RecordZone       string      `json:"record_zone" mapstructure:"record_zone"`
	RecordName       string      `json:"record_name" mapstructure:"record_name"`
	RecordType       string      `json:"record_type" mapstructure:"record_type"`
	RecordData       string      `json:"record_data" mapstructure:"record_data"`
	RecordAux        int64       `json:"record_aux" mapstructure:"record_aux"`
	RecordId         StringOrInt `json:"record_id" mapstructure:"record_id"`
	RecordChangeable string      `json:"record_changeable" mapstructure:"record_changeable"`
	RecordDeleteable string      `json:"record_deleteable" mapstructure:"record_deleteable"`
}

type AddDNSAPIResponse struct {
	Response AddDNSResponse `json:"Response" mapstructure:"Response"`
}

type AddDNSResponse struct {
	KasFloodDelay float64 `json:"KasFloodDelay" mapstructure:"KasFloodDelay"`
	ReturnInfo    string  `json:"ReturnInfo" mapstructure:"ReturnInfo"`
	ReturnString  string  `json:"ReturnString" mapstructure:"ReturnString"`
}

type UpdateDNSAPIResponse struct {
	Response UpdateDNSResponse `json:"Response" mapstructure:"Response"`
}

type UpdateDNSResponse struct {
	KasFloodDelay float64 `json:"KasFloodDelay" mapstructure:"KasFloodDelay"`
	ReturnInfo    string  `json:"ReturnInfo" mapstructure:"ReturnInfo"`
	ReturnString  string  `json:"ReturnString" mapstructure:"ReturnString"`
}

type DeleteDNSAPIResponse struct {
	Response DeleteDNSResponse `json:"Response" mapstructure:"Response"`
}

type DeleteDNSResponse struct {
	KasFloodDelay float64 `json:"KasFloodDelay" mapstructure:"KasFloodDelay"`
	ReturnInfo    string  `json:"ReturnInfo" mapstructure:"ReturnInfo"`
	ReturnString  string  `json:"ReturnString" mapstructure:"ReturnString"`
}

type GetDDNSUserAPIResponse struct {
	Response GetDDNSUserResponse `json:"Response" mapstructure:"Response"`
}

type GetDDNSUserResponse struct {
	KasFloodDelay float64             `json:"KasFloodDelay" mapstructure:"KasFloodDelay"`
	ReturnInfo    []GetDDNSReturnInfo `json:"ReturnInfo" mapstructure:"ReturnInfo"`
	ReturnString  string              `json:"ReturnString"`
}

type GetDDNSReturnInfo struct {
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

// StringOrInt is a helper type that can unmarshal JSON values
// that may be either a string or an integer. The value is stored
// as a string internally.
type StringOrInt string

func (s *StringOrInt) UnmarshalJSON(b []byte) error {
	// Handle quoted string
	if len(b) > 0 && b[0] == '"' {
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}
		*s = StringOrInt(str)
		return nil
	}
	// Handle integer number
	var num int64
	if err := json.Unmarshal(b, &num); err == nil {
		*s = StringOrInt(fmt.Sprintf("%d", num))
		return nil
	}
	// Handle null gracefully
	if string(b) == "null" {
		*s = ""
		return nil
	}
	return fmt.Errorf("StringOrInt: unsupported JSON value: %s", string(b))
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
