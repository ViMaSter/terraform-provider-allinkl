package allinkl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
)

const apiEndpoint = "https://kasapi.kasserver.com/soap/KasApi.php"

type Authentication interface {
	Authentication(ctx context.Context, sessionLifetime int, sessionUpdateLifetime bool) (string, error)
}

// Client a KAS server client.
type Client struct {
	identifier *Identifier
	baseURL    string
	HTTPClient *http.Client
}

var floodTime time.Time
var muFloodTime sync.Mutex

func NewClient(username string, kasAuthType string, kasAuthData string) *Client {
	return &Client{
		identifier: NewIdentifier(username, kasAuthType, kasAuthData),
		baseURL:    apiEndpoint,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

}
func (c *Client) GetDDNSUser(ctx context.Context, ddnsLogin string) (ReturnInfo, error) {
	requestParams := map[string]string{"ddns_login": ddnsLogin}

	credential, err := c.identifier.Authentication(ctx)
	if err != nil {
		var empty ReturnInfo
		return empty, err
	}

	ctx = WithContext(ctx, credential)

	req, err := c.newRequest(ctx, "get_ddnsusers", requestParams)
	if err != nil {
		var empty ReturnInfo
		return empty, err
	}
	var g GetDDNSUserAPIResponse
	err = c.do(req, &g)
	if err != nil {
		var empty ReturnInfo
		return empty, err
	}

	if len(g.Response.ReturnInfo) != 1 {
		var empty ReturnInfo
		return empty, fmt.Errorf("expected exactly 1 DDNS user, got %d", len(g.Response.ReturnInfo))
	}

	c.updateFloodTime(g.Response.KasFloodDelay)
	return g.Response.ReturnInfo[0], nil
}

func (c *Client) AddDDNSUser(ctx context.Context, record DDNSRequest) (string, error) {
	credential, err := c.identifier.Authentication(ctx)
	if err != nil {
		return "", err
	}

	ctx = WithContext(ctx, credential)

	req, err := c.newRequest(ctx, "add_ddnsuser", record)
	if err != nil {
		return "", err
	}
	var g AddDDNSUserAPIResponse
	err = c.do(req, &g)
	if err != nil {
		return "", err
	}
	c.updateFloodTime(g.Response.KasFloodDelay)
	return g.Response.ReturnInfo, nil
}

func (c *Client) UpdateDDNSUser(ctx context.Context, record DDNSUpdateRequest) (string, error) {
	credential, err := c.identifier.Authentication(ctx)
	if err != nil {
		return "", err
	}

	ctx = WithContext(ctx, credential)

	req, err := c.newRequest(ctx, "update_ddnsuser", record)
	if err != nil {
		return "", err
	}
	var g AddDDNSUserAPIResponse
	err = c.do(req, &g)
	if err != nil {
		return "", err
	}
	c.updateFloodTime(g.Response.KasFloodDelay)
	return g.Response.ReturnString, nil
}

func (c *Client) DeleteDDNSUser(ctx context.Context, dyndnsLogin string) (string, error) {
	credential, err := c.identifier.Authentication(ctx)
	if err != nil {
		return "", err
	}

	ctx = WithContext(ctx, credential)

	requestParams := map[string]string{"dyndns_login": dyndnsLogin}
	req, err := c.newRequest(ctx, "delete_ddnsuser", requestParams)
	if err != nil {
		return "", err
	}
	var g DeleteDDNSUserAPIResponse
	err = c.do(req, &g)
	if err != nil {
		return "", err
	}
	c.updateFloodTime(g.Response.KasFloodDelay)
	return g.Response.ReturnInfo, nil
}

func (c *Client) newRequest(ctx context.Context, action string, requestParams any) (*http.Request, error) {
	ar := KasRequest{
		Login:         c.identifier.login,
		AuthType:      "session",
		AuthData:      getToken(ctx),
		Action:        action,
		RequestParams: requestParams,
	}
	body, err := json.Marshal(ar)
	if err != nil {
		return nil, fmt.Errorf("failed to create request JSON body: %w", err)
	}
	payload := []byte(strings.TrimSpace(fmt.Sprintf(kasAPIEnvelope, body)))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}
	return req, nil
}

func (c *Client) do(req *http.Request, result any) error {
	muFloodTime.Lock()
	sleepDuration := time.Until(floodTime)
	if sleepDuration > 0 {
		time.Sleep(sleepDuration)
	}
	muFloodTime.Unlock()
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return NewHTTPDoError(req, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return NewUnexpectedResponseStatusCodeError(req, resp)
	}
	envlp, err := decodeXML[KasAPIResponseEnvelope](resp.Body)
	if err != nil {
		return err
	}
	if envlp.Body.Fault != nil {
		return envlp.Body.Fault
	}
	raw := getValue(envlp.Body.KasAPIResponse.Return)
	err = mapstructure.Decode(raw, result)
	if err != nil {
		return fmt.Errorf("response struct decode: %w", err)
	}
	return nil
}

func (c *Client) updateFloodTime(delay float64) {
	muFloodTime.Lock()
	floodTime = time.Now().Add(time.Duration(delay * float64(time.Second*5)))
	muFloodTime.Unlock()
}

func getValue(item *Item) any {
	switch {
	case item.Raw != "":
		v, _ := strconv.ParseBool(item.Raw)
		return v
	case item.Text != "":
		switch item.Type {
		case "xsd:string":
			return item.Text
		case "xsd:float":
			v, _ := strconv.ParseFloat(item.Text, 64)
			return v
		case "xsd:int":
			v, _ := strconv.ParseInt(item.Text, 10, 64)
			return v
		default:
			return item.Text
		}
	case item.Value != nil:
		return getValue(item.Value)
	case len(item.Items) > 0 && item.Type == "SOAP-ENC:Array":
		var v []any
		for _, i := range item.Items {
			v = append(v, getValue(i))
		}
		return v
	case len(item.Items) > 0:
		v := map[string]any{}
		for _, i := range item.Items {
			v[getKey(i)] = getValue(i)
		}
		return v
	default:
		return ""
	}
}

func getKey(item *Item) string {
	if item.Key == nil {
		return ""
	}
	return item.Key.Text
}
