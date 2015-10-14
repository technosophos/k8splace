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

	app.Commands = []cli.Command{
		{
			Name:   "get",
			Usage:  "Get a package from k8splace",
			Action: get,
		},
		{
			Name:   "install",
			Usage:  "Install a package into a Kubernetes cluster",
			Action: install,
		},
		{
			Name:   "push",
			Usage:  "Push an updated package to k8splace",
			Action: push,
		},
	}

	app.Run(os.Args)
}

func die(err error) {
	m := "{{.Red}}[BOO!]{{.Default}} " + err.Error()
	fmt.Println(pretty.Colorize(m))
	os.Exit(1)
}

func get(c *cli.Context) {
	pkg := a1(c, "No package name given")
	info("Getting %s", pkg)
	cwd, _ := os.Getwd()
	ftw("Installed %q into %q", pkg, path.Join(cwd, pkg))
}

func install(c *cli.Context) {
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
