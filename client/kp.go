package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/codegangsta/cli"
	pretty "github.com/deis/pkg/prettyprint"
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
	}

	app.Run(os.Args)
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

	if len(p.Releases) == 0 {
		info("There are no releases of %q. Aborting.", pkg)
		os.Exit(2)
	}

	rel := p.Releases[0]
	info("Newest Version: %s", rel.Version)

	ftw("Installed %s %s into %q", p.Name, rel.Version, path.Join(wd, pkg))
}

func install(c *cli.Context) {
	ensureHome(c)
	pkg := a1(c, "No package name given")
	info("Getting %s", pkg)
	cwd, _ := os.Getwd()
	ftw("Installed %q into %q", pkg, path.Join(cwd, pkg))
	info("Uploading %s to Kubernetes", pkg)
}

func push(c *cli.Context) {
	pkg := a1(c, "No package name given")
	info("Pushing %q to K8sPlace", pkg)
}

func a1(c *cli.Context, msg string) string {
	a := c.Args()
	if len(a) < 1 {
		die(errors.New(msg))
	}
	return a[0]
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
