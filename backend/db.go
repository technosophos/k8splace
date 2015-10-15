package backend

import (
	"github.com/Masterminds/cookoo"
	"github.com/technosophos/k8splace/model"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const DSName = "db"

type Datasource struct {
	session *mgo.Session
	dbName  string
	db      *mgo.Database
}

func NewDS(host, db string) (*Datasource, error) {
	s, err := mgo.Dial(host)
	if err != nil {
		return nil, err
	}

	return &Datasource{
		session: s,
		dbName:  db,
		db:      s.DB(db),
	}, nil
}

// C gets a named collection.
func (d *Datasource) C(name string) *mgo.Collection {
	return d.db.C(name)
}

func (d *Datasource) Close() {
	d.session.Close()
}

func ds(c cookoo.Context) *Datasource {
	return c.Datasource(DSName).(*Datasource)
}

// Package gets a package.
//
// Params:
//	- name (string): Package name
// Returns:
// 	- *model.Package
func Package(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	name := p.Get("name", "").(string)

	var pkg model.Package
	err := ds(c).C("packages").Find(bson.M{"name": name}).One(&pkg)

	return &pkg, err
}

// AddPackage adds a package to the service.
//
// Params:
// 	- pkg (*model.Package): The package to add
// Returns:
//  - pkg (*model.Package): The package, modified
func AddPackage(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	pkg := p.Get("pkg", nil).(*model.Package)

	err := ds(c).C("packages").Insert(pkg)
	return pkg, err
}
