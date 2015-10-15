package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"

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
					{Name: "dest", From: "cxt:prototype"},
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
				Fn:   backend.Packages,
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
				Name: "-",
				Fn:   cookoo.AddToContext,
				Using: []cookoo.Param{
					{Name: "prototype", DefaultValue: &model.Package{}},
				},
			},
			cookoo.Include{"@json"},
			cookoo.Cmd{
				Name: "res",
				Fn:   backend.AddPackage,
				Using: []cookoo.Param{
					{Name: "pkg", From: "cxt:json"},
				},
			},
			cookoo.Include{"@out"},
		},
	})
	reg.AddRoute(cookoo.Route{
		Name: "POST /package/*",
		Help: "Create a package release",
		Does: []cookoo.Task{
			//cookoo.Cmd{
			//Name: "pkg",
			//Fn:   backend.Package,
			//Using: []cookoo.Param{
			//{Name: "name", From: "path:1"},
			//},
			//},
			cookoo.Cmd{
				Name: "-",
				Fn:   cookoo.AddToContext,
				Using: []cookoo.Param{
					{Name: "prototype", DefaultValue: &model.Package{}},
				},
			},
			cookoo.Include{"@json"},
			cookoo.Cmd{
				Name: "res",
				Fn:   backend.AddRelease,
				Using: []cookoo.Param{
					{Name: "pkg", From: "cxt:json"},
				},
			},
			cookoo.Include{"@out"},
		},
	})
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

	// Sort of a dumb way to clone the value of a pointer.
	tt := reflect.Indirect(reflect.ValueOf(dest)).Type()
	out := reflect.New(tt).Interface()

	return out, json.Unmarshal(data, out)
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
