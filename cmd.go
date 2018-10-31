package gocmd

import (
	"bytes"
	"os/exec"
	"syscall"
	"time"
)

type CMD struct {
	cmd *exec.Cmd
	binPath string
	params map[string]string
	exitStatus int
	stderr bytes.Buffer
	stdout bytes.Buffer
	running bool
}

func NewCmd() *CMD {
	cmd := &CMD{}
	cmd.params = make(map[string]string)
	cmd.exitStatus = 0
	cmd.running = false

	return cmd
}

func (c *CMD) SetBinPath(path string) {
	c.binPath = path
}

func (c *CMD) SetParam(param, value string) {
	c.params[param] = value
}

func (c *CMD) SetParams(params map[string]string) {
	for k, v := range params {
		c.SetParam(k, v)
	}
}

func (c *CMD) GetStdout() string {
	return c.stdout.String()
}

func (c *CMD) GetStderr() string {
	return c.stderr.String()
}

func (c *CMD) GetExitStatus() int {
	return c.exitStatus
}

func (c *CMD) IsRunning() bool {
	return c.running
}

func (c *CMD) Start() error {
	var execCmdArgs []string
	for k, v := range c.params {
		execCmdArgs = append(execCmdArgs, k, v)
	}

	c.cmd = exec.Command(c.binPath, execCmdArgs...)

	c.cmd.Stdout = &c.stdout
	c.cmd.Stderr = &c.stderr

	err := c.cmd.Start()
	if err != nil {
		return err
	}

	c.running = true

	go func(c *CMD) {
		err := c.cmd.Wait()
		if err != nil {
			exiterr, ok := err.(*exec.ExitError)
			if ok {
				status := exiterr.Sys().(syscall.WaitStatus)
				c.exitStatus = status.ExitStatus()
			} else {
				c.exitStatus = 1
			}
		} else {
			status := c.cmd.ProcessState.Sys().(syscall.WaitStatus)
			c.exitStatus = status.ExitStatus()
		}

		c.running = false
	}(c)

	return nil
}

func (c *CMD) Run() (int, string, string, error) {
	err := c.Start()
	if err != nil {
		return c.GetExitStatus(), c.GetStdout(), c.GetStderr(), err
	}

	for c.IsRunning() {
		time.Sleep(time.Microsecond)
	}

	return c.GetExitStatus(), c.GetStdout(), c.GetStderr(), nil
}