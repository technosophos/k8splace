package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/codegangsta/cli"
	pretty "github.com/deis/pkg/prettyprint"
	"github.com/technosophos/k8splace/model"
)

const version = "0.0.1"

func main() {
	app := cli.NewApp()
	app.Name = "kp"
	app.Usage = "Work with the k8splace hub"
	app.Version = version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "host",
			Value:  "http://localhost:8080",
			Usage:  "The k8splace server URL",
			EnvVar: "K8SPLACE_URL",
		},
		cli.StringFlag{
			Name:   "homedir",
			Value:  "$HOME/.k8splace",
			Usage:  "The location of your k8splace archive",
			EnvVar: "K8SPLACE_HOME",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "get",
			Usage:  "get packge:name - Get a package from k8splace",
			Action: get,
		},
		{
			Name:   "install",
			Usage:  "install package:name - Install a package into a Kubernetes cluster",
			Action: install,
		},
		{
			Name:   "push",
			Usage:  "push package:name - Push an updated package to k8splace",
			Action: push,
		},
		{
			Name:   "create",
			Usage:  "create package:name - Create a new packages",
			Action: create,
		},
		{
			Name:   "list",
			Usage:  "List all packages",
			Action: list,
		},
	}

	app.Run(os.Args)
}

func list(c *cli.Context) {
	h := NewClient(c.GlobalString("host"))
	items, err := h.List()
	if err != nil {
		die(err)
	}
	for _, i := range items.Results {
		fmt.Printf("\t%s - %s\n", i.Name, i.Description)
	}
}

func get(c *cli.Context) {
	wd := ensureHome(c)
	pkg := a1(c, "No package name given")
	info("Getting %s", pkg)

	h := NewClient(c.GlobalString("host"))
	p, err := h.Get(pkg)
	if err != nil {
		die(err)
	}

	pdir := path.Join(wd, p.Name)

	if err := os.MkdirAll(pdir, 0755); err != nil {
		die(err)
	}

	if len(p.Releases) == 0 {
		info("There are no releases of %q.", pkg)
		ftw("Installed empty package %s into %q", p.Name, path.Join(wd, pdir))
		return
	}

	rel := p.Releases[0]
	info("Newest Version: %s", rel.Version)

	ftw("Installed %s %s into %q", p.Name, rel.Version, pdir)
}

func install(c *cli.Context) {
	wd := ensureHome(c)
	pkg := a1(c, "No package name given")
	fname := path.Join(wd, pkg)
	if _, err := os.Stat(fname); err != nil {
		get(c)
	}

	info("Uploading %s to Kubernetes", pkg)
	paths, _ := filepath.Glob(filepath.Join(fname, "*.yaml"))
	for _, ff := range paths {
		cmd := exec.Command("kubectl", "create", "-f", ff)
		if out, err := cmd.CombinedOutput(); err != nil {
			info("%s: %s", out, err)
		}
	}
	paths, _ = filepath.Glob(filepath.Join(fname, "*.json"))
	for _, ff := range paths {
		cmd := exec.Command("kubectl", "create", "-f", ff)
		if out, err := cmd.CombinedOutput(); err != nil {
			info("%s: %s", out, err)
		}
	}
}

func create(c *cli.Context) {
	fname := a1(c, "No JSON file given")
	h := NewClient(c.GlobalString("host"))
	_, err := h.CreatePackage(fname)
	if err != nil {
		die(err)
	}
	ftw("Created project")
}

func push(c *cli.Context) {
	wd := ensureHome(c)
	pkg := a1(c, "No package name given")
	ver := a2(c, "Version is required")
	h := NewClient(c.GlobalString("host"))

	// Grab the latest version of the package from the server
	p, err := h.Get(pkg)
	if err != nil {
		die(err)
	}

	// Build a release
	r := &model.Release{
		Version:   ver,
		Author:    p.Author,
		Date:      time.Now(),
		PackageId: p.ID,
		Manifests: makeManifests(path.Join(wd, pkg)),
	}

	// Add it
	p.Releases = append([]*model.Release{r}, p.Releases...)

	info("Pushing %s %s to K8sPlace", pkg, ver)
	//d, _ := json.Marshal(p)
	//fmt.Println(string(d))
	if err := h.Update(p); err != nil {
		die(err)
	}
	ftw("Updated %s", p.Name)
}

func makeManifests(wd string) []*model.Manifest {
	//m := map[string]string{}
	m := []*model.Manifest{}

	f, err := os.Open(wd)
	if err != nil {
		die(err)
	}

	files, err := f.Readdirnames(0)
	if err != nil {
		die(err)
	}

	for _, fname := range files {
		// Read the contents
		data, err := ioutil.ReadFile(path.Join(wd, fname))
		if err != nil {
			info("Skipping file %s because %q", fname, err)
			continue
		}
		//m[fname] = string(data)
		mm := &model.Manifest{
			Name: fname,
			Data: string(data),
		}
		m = append(m, mm)
	}

	return m
}

func a1(c *cli.Context, msg string) string {
	a := c.Args()
	if len(a) < 1 {
		die(errors.New(msg))
	}
	return a[0]
}

func a2(c *cli.Context, msg string) string {
	a := c.Args()
	if len(a) < 2 {
		die(errors.New(msg))
	}
	return a[1]
}

func info(msg string, args ...interface{}) {
	t := fmt.Sprintf(msg, args...)
	m := "{{.Yellow}}[INFO]{{.Default}} " + t
	fmt.Println(pretty.Colorize(m))
}

func ftw(msg string, args ...interface{}) {
	t := fmt.Sprintf(msg, args...)
	m := "{{.Green}}[YAY!]{{.Default}} " + t
	fmt.Println(pretty.Colorize(m))
}

func ensureHome(c *cli.Context) string {
	wd := c.GlobalString("homedir")
	wd = os.ExpandEnv(wd)
	if _, err := os.Stat(wd); err != nil {
		info("Attempting to create dir %q", wd)
		if err := os.MkdirAll(wd, 0755); err != nil {
			die(err)
		}
		ftw("Created")
	}
	return wd
}

func die(err error) {
	m := "{{.Red}}[BOO!]{{.Default}} " + err.Error()
	fmt.Println(pretty.Colorize(m))
	os.Exit(1)
}
