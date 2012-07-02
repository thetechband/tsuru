package service

import (
	"encoding/json"
	"errors"
	"github.com/timeredbull/tsuru/api/app"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	endpoint string
}

func (c *Client) issue(path, method string, params map[string][]string) (*http.Response, error) {
	v := url.Values(params)
	body := strings.NewReader(v.Encode())
	url := strings.TrimRight(c.endpoint, "/") + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func (c *Client) jsonFromResponse(resp *http.Response) (env map[string]string, err error) {
	defer resp.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &env)
	return
}

func (c *Client) Create(instance *ServiceInstance) (envVars map[string]string, err error) {
	var resp *http.Response
	params := map[string][]string{
		"name": []string{instance.Name},
	}
	if resp, err = c.issue("/resources", "POST", params); err == nil {
		return c.jsonFromResponse(resp)
	}
	return
}

func (c *Client) Destroy(instance *ServiceInstance) (err error) {
	var resp *http.Response
	if resp, err = c.issue("/resources/"+instance.Name, "DELETE", nil); err == nil && resp.StatusCode > 299 {
		err = errors.New("Failed to destroy the instance: " + instance.Name)
	}
	return err
}

func (c *Client) Bind(instance *ServiceInstance, app *app.App, serviceHost string) (envVars map[string]string, err error) {
	var resp *http.Response
	params := map[string][]string{
		"hostname":     []string{app.Units[0].Ip},
		"service_host": []string{serviceHost},
	}
	if resp, err = c.issue("/resources/"+instance.Name, "POST", params); err == nil {
		return c.jsonFromResponse(resp)
	}
	return
}
