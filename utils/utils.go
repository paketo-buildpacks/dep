package utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type Command struct {}

func (c *Command) Run(dir, bin string, args ...string) (string, error) {
	logs := &bytes.Buffer{}
	err := c.run(dir, []string{}, io.MultiWriter(os.Stdout, logs), io.MultiWriter(os.Stderr, logs), bin, args...)
	return logs.String(), err
}

func (c *Command) RunSilent(dir, bin string, args ...string) (string, error) {
	logs := &bytes.Buffer{}
	err := c.run(dir, []string{}, io.MultiWriter(ioutil.Discard, logs), io.MultiWriter(ioutil.Discard, logs), bin, args...)
	return logs.String(), err
}

func (c *Command) CustomRun(dir string, env []string, out, err io.Writer, bin string, args ...string) error {
	return c.run(dir, env, out, err, bin, args...)
}

func (c *Command) run(dir string, env []string, out, err io.Writer, bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Stdout = out
	cmd.Stderr = err
	cmd.Env = append(os.Environ(), env...)
	return cmd.Run()
}
