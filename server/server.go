package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/cookoo/log"
	"github.com/Masterminds/cookoo/web"
	"github.com/technosophos/k8splace/backend"
	"github.com/technosophos/k8splace/model"
)

func main() {
	reg, route, c := cookoo.Cookoo()

	ds, err := backend.NewDS("localhost", "k8e")
	if err != nil {
		log.Critf(c, "Shutting down: %s", err)
		return
	}
	defer ds.Close()

	c.AddDatasource("db", ds)

	routes(reg)
	web.Serve(reg, route, c)
}

func routes(reg *cookoo.Registry) {
	reg.AddRoute(cookoo.Route{
		Name: "@json",
		Help: "Parse incoming JSON out of HTTP body",
		Does: []cookoo.Task{
			cookoo.Cmd{
				Name: "data",
				Fn:   readBody,
			},
			cookoo.Cmd{
				Name: "json",
				Fn:   fromJSON,
				Using: []cookoo.Param{
					{Name: "data", From: "cxt:data"},
					{Name: "dest", From: "cxt:dest"},
				},
			},
		},
	})
	reg.AddRoute(cookoo.Route{
		Name: "@out",
		Help: "Serialize JSON and write it to http.Response",
		Does: []cookoo.Task{
			cookoo.Cmd{
				Name: "content",
				Fn:   toJSON,
				Using: []cookoo.Param{
					{Name: "o", From: "cxt:res"},
				},
			},
			cookoo.Cmd{
				Name: "flush",
				Fn:   web.Flush,
				Using: []cookoo.Param{
					{Name: "contentType", DefaultValue: "application/json"},
					{Name: "content", From: "cxt:content"},
				},
			},
		},
	})
	reg.AddRoute(cookoo.Route{
		Name: "GET /package",
		Help: "List packages",
		Does: []cookoo.Task{
			cookoo.Cmd{
				Name: "res",
				Fn:   Packages,
			},
			cookoo.Include{"@out"},
		},
	})
	reg.AddRoute(cookoo.Route{
		Name: "GET /package/*",
		Help: "Get an individual package",
		Does: []cookoo.Task{
			cookoo.Cmd{
				Name: "res",
				Fn:   backend.Package,
				Using: []cookoo.Param{
					{Name: "name", From: "path:1"},
				},
			},
			cookoo.Include{"@out"},
		},
	})
	reg.AddRoute(cookoo.Route{
		Name: "POST /package",
		Help: "Create a package",
		Does: []cookoo.Task{
			cookoo.Cmd{
				Name: "res",
				Fn:   backend.AddPackage,
			},
			cookoo.Include{"@out"},
		},
	})
	reg.AddRoute(cookoo.Route{
		Name: "POST /package/*",
		Help: "Create a package release",
		Does: []cookoo.Task{
			cookoo.Cmd{
				Name: "res",
				Fn:   AddRelease,
			},
			cookoo.Include{"@out"},
		},
	})
}

// Packages lists the packages
//
// Returns:
//	- *model.Packages
func Packages(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	return &model.Results{
		Count:  3,
		Offset: 0,
		Total:  3,
		Results: []*model.Package{
			dummyPackage(1, "deis:postgres"),
			dummyPackage(2, "deis:riak"),
			dummyPackage(3, "technosophos:blog"),
		},
	}, nil
}

func dummyRelease(id, parent int, ver string) *model.Release {
	return &model.Release{
		ID:          id,
		Version:     ver,
		Description: "Fix all the things!",
		Author:      "technosophos@github.com",
		Date:        time.Now(),
		Rating:      4.7,
		Manifests:   []interface{}{},
		PackageId:   parent,
	}
}

func dummyPackage(id int, name string) *model.Package {
	return &model.Package{
		ID:          id,
		Name:        name,
		Description: "Postgres using Governor and Etcd for HA",
		Readme: `# Postgres
		Run postgres in Kubernetes.
		`,
		Author:       "Jack, Rimus, and Matt",
		CreationDate: time.Now().Add(-144 * time.Hour),
		Rating:       5.0,
		LastUpdated:  time.Now().Add(-5 * time.Minute),
		Releases: []*model.Release{
			dummyRelease(300+id, id, "1.3.0"),
			dummyRelease(200+id, id, "1.2.4"),
			dummyRelease(100+id, id, "1.2.3"),
		},
	}
}

var PkgNotFound = errors.New("Package not found")

// GetPackage gets a package by ID
//
// Params:
//	- id (string)
//
// Returns:
//  - *model.Package
func GetPackage(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	pname := p.Get("id", "").(string)

	if pname != "deis:postgres" {
		log.Warnf(c, "Not found: %q", pname)
		// FIXME: This should not be a 500
		return nil, PkgNotFound
	}
	return dummyPackage(1, "deis:postgres"), nil
}

// AddPackage creates a new package
//
// Params:
//
// Returns:
//
func AddPackage(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	return dummyPackage(1, "deis:postgres"), nil
}

// AddRelease adds a release to a package
//
// Params:
//
// Returns:
//
func AddRelease(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	return dummyPackage(1, "deis:postgres"), nil
}

// toJSON marshals an object into JSON
//
// Params:
//	o (interface{}): an object to marshal
// Returns:
//  []byte
func toJSON(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	return json.Marshal(p.Get("o", ""))
}

// fromJSON parses JSON into the named interface.
//
// Params:
// 	- data []byte: the raw data
//	- dest interface{}: The destination object
// Returns:
//
func fromJSON(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	dest := p.Get("dest", nil)
	data := p.Get("data", []byte{}).([]byte)

	return &dest, json.Unmarshal(data, &dest)
}

// readBody reads the body of an http request
//
// Params:
//
// Returns:
// 	[]byte
func readBody(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	r := c.Get("http.Request", nil).(*http.Request)
	return ioutil.ReadAll(r.Body)
}
