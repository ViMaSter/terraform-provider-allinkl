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

type DNSRequest struct {
	// RecordId the ID of the resource record
	RecordId string `json:"record_id,omitempty"`
	// ZoneHost the zone in question (must be a FQDN).
	ZoneHost string `json:"zone_host"`
	// RecordType the TYPE of the resource record (MX, A, AAAA etc.).
	RecordType string `json:"record_type"`
	// RecordName the NAME of the resource record.
	RecordName string `json:"record_name"`
	// RecordData the DATA of the resource record.
	RecordData string `json:"record_data"`
	// RecordAux the AUX of the resource record.
	RecordAux int `json:"record_aux"`
}

// ---

type GetDNSSettingsAPIResponse struct {
	Response GetDNSSettingsResponse `json:"Response" mapstructure:"Response"`
}

type GetDNSSettingsResponse struct {
	KasFloodDelay float64      `json:"KasFloodDelay" mapstructure:"KasFloodDelay"`
	ReturnInfo    []ReturnInfo `json:"ReturnInfo" mapstructure:"ReturnInfo"`
	ReturnString  string       `json:"ReturnString"`
}

type ReturnInfo struct {
	ID         any    `json:"record_id,omitempty" mapstructure:"record_id"`
	ZoneHost   string `json:"record_zone,omitempty" mapstructure:"record_zone"`
	RecordName string `json:"record_name,omitempty" mapstructure:"record_name"`
	RecordType string `json:"record_type,omitempty" mapstructure:"record_type"`
	RecordData string `json:"record_data,omitempty" mapstructure:"record_data"`
	Changeable string `json:"record_changeable,omitempty" mapstructure:"record_changeable"`
	RecordAux  int    `json:"record_aux,omitempty" mapstructure:"record_aux"`
}

type AddDNSSettingsAPIResponse struct {
	Response AddDNSSettingsResponse `json:"Response" mapstructure:"Response"`
}

type AddDNSSettingsResponse struct {
	KasFloodDelay float64 `json:"KasFloodDelay" mapstructure:"KasFloodDelay"`
	ReturnInfo    string  `json:"ReturnInfo" mapstructure:"ReturnInfo"`
	ReturnString  string  `json:"ReturnString" mapstructure:"ReturnString"`
}

type DeleteDNSSettingsAPIResponse struct {
	Response DeleteDNSSettingsResponse `json:"Response"`
}

type DeleteDNSSettingsResponse struct {
	KasFloodDelay float64 `json:"KasFloodDelay"`
	ReturnInfo    bool    `json:"ReturnInfo"`
	ReturnString  string  `json:"ReturnString"`
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
