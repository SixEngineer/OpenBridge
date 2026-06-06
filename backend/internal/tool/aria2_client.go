package tool

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// aria2c --enable-rpc --rpc-listen-all

type Aria2Client struct {
	rpcURL string
	secret string
	client *http.Client
}

type aria2Request struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params,omitempty"`
}

type aria2Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type aria2Response struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result"`
	Error   *aria2Error `json:"error"`
}

func NewAria2Client(rpcURL, secret string) *Aria2Client {
	return &Aria2Client{
		rpcURL: rpcURL,
		secret: secret,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *Aria2Client) AddURI(uri string) (string, error) {
	return a.AddURIWithOptions(uri, nil)
}

func (a *Aria2Client) AddURIWithOptions(uri string, options map[string]interface{}) (string, error) {
	params := []interface{}{}
	if a.secret != "" {
		params = append(params, "token:"+a.secret)
	}
	params = append(params, []string{uri})
	if options != nil {
		params = append(params, options)
	}

	result, err := a.call("aria2.addUri", params)
	if err != nil {
		return "", err
	}

	gid, ok := result.(string)
	if !ok {
		return "", errors.New("aria2 addUri response invalid")
	}
	return gid, nil
}

func (a *Aria2Client) TellStatus(gid string) (map[string]interface{}, error) {
	params := []interface{}{}
	if a.secret != "" {
		params = append(params, "token:"+a.secret)
	}
	params = append(params, gid)

	result, err := a.call("aria2.tellStatus", params)
	if err != nil {
		return nil, err
	}

	status, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("aria2 tellStatus response invalid")
	}
	return status, nil
}

func (a *Aria2Client) Remove(gid string) (string, error) {
	params := []interface{}{}
	if a.secret != "" {
		params = append(params, "token:"+a.secret)
	}
	params = append(params, gid)

	result, err := a.call("aria2.remove", params)
	if err != nil {
		return "", err
	}

	removed, ok := result.(string)
	if !ok {
		return "", errors.New("aria2 remove response invalid")
	}
	return removed, nil
}

func (a *Aria2Client) call(method string, params []interface{}) (interface{}, error) {
	reqBody := aria2Request{
		Jsonrpc: "2.0",
		ID:      "openbridge",
		Method:  method,
		Params:  params,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", a.rpcURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp aria2Response
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, errors.New(rpcResp.Error.Message)
	}
	return rpcResp.Result, nil
}
