package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
	"mize.app/logger"
)

type NetworkController struct {
	BaseUrl    string
	HttpClinet *http.Client
}

func (network *NetworkController) InitialiseNetworkClient() {
	network.HttpClinet = &http.Client{}
}

func (network *NetworkController) Get(path string, headers *map[string]string, params *map[string]string) (*string, error) {
	if network.HttpClinet == nil {
		network.InitialiseNetworkClient()
	}
	req, err := http.NewRequest("GET", network.BaseUrl+path, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	setHeaders(headers, req)
	setParams(params, req)
	res, err := network.HttpClinet.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	res_body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	res_json := string(res_body)
	fmt.Println(res_json)
	fmt.Println(res.Status)
	return &res_json, nil
}

func (network *NetworkController) Post(path string, headers *map[string]string, body *map[string]interface{}, params *map[string]string) (*string, error) {
	if network.HttpClinet == nil {
		network.InitialiseNetworkClient()
	}
	parsed_body, err := json.Marshal(body)
	if err != nil {
		logger.Error(errors.New("error converting body to JSON"), zap.Error(err))
		return nil, err
	}
	req, err := http.NewRequest("POST", network.BaseUrl+path, bytes.NewBuffer(parsed_body))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	setHeaders(headers, req)
	setParams(params, req)
	defer req.Body.Close()
	res, err := network.HttpClinet.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	res_body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var t interface{}
	json.Unmarshal(res_body, &t)
	res_json := string(res_body)
	return &res_json, nil
}

func setHeaders(headers *map[string]string, req *http.Request) {
	if headers == nil {
		return
	}
	for k := range *headers {
		req.Header.Add(k, (*headers)[k])
	}
}

func setParams(params *map[string]string, req *http.Request) {
	if params == nil {
		return
	}
	q := req.URL.Query()
	for k := range *params {
		q.Add(k, (*params)[k])
	}
	req.URL.RawQuery = q.Encode()
}
