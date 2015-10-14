package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/technosophos/k8splace/model"
)

type Client struct {
	addr string
}

func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

func (c *Client) Get(pkg string) (*model.Package, error) {
	r, err := http.Get(c.addr + "/package/" + pkg)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != 200 {
		return nil, fmt.Errorf("Server responded %s", r.Status)
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	p := &model.Package{}
	err = json.Unmarshal(data, p)
	return p, err
}
