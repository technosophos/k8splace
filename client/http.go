package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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

func (c *Client) List() (*model.Results, error) {
	r, err := http.Get(c.addr + "/package")
	if err != nil {
		return nil, err
	}
	if r.StatusCode < 200 || r.StatusCode > 299 {
		return nil, fmt.Errorf("Server responded %s", r.Status)
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	res := &model.Results{}
	return res, json.Unmarshal(b, res)
}

func (c *Client) Update(pkg *model.Package) error {
	u := c.addr + "/package/" + pkg.Name
	data, err := json.Marshal(pkg)
	if err != nil {
		return err
	}

	r, err := http.Post(u, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	if r.StatusCode < 200 || r.StatusCode > 299 {
		return fmt.Errorf("Server responded %s", r.Status)
	}

	return nil
}

func (c *Client) CreatePackage(fname string) (*model.Package, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	u := c.addr + "/package"
	r, err := http.Post(u, "application/json", f)
	if err != nil {
		return nil, err
	}
	if r.StatusCode < 200 || r.StatusCode > 299 {
		return nil, fmt.Errorf("Server responded %s", r.Status)
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	v := &model.Package{}
	return v, json.Unmarshal(b, v)
}
