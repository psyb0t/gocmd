package gocmd

// Easily execute shell commands
// Wrapper for os/exec

import (
	"bytes"
	"os/exec"
	"syscall"
	"time"
	"os"
)

// This type contains the command instructions and
// return values
type CMD struct {
	cmd *exec.Cmd
	binPath string
	params map[string]string
	exitStatus int
	stderr bytes.Buffer
	stdout bytes.Buffer
	running bool
}

// Returns a new CMD struct
func NewCmd() *CMD {
	cmd := &CMD{}
	cmd.params = make(map[string]string)
	cmd.exitStatus = 0
	cmd.running = false

	return cmd
}

// Set the path of the executable to be run
func (c *CMD) SetBinPath(path string) {
	c.binPath = path
}

// Set one parameter for the executable
func (c *CMD) SetParam(param, value string) {
	c.params[param] = value
}

// Set multiple parameters for the executable
func (c *CMD) SetParams(params map[string]string) {
	for k, v := range params {
		c.SetParam(k, v)
	}
}

// Return the STDOUT string
func (c *CMD) GetStdout() string {
	return c.stdout.String()
}

// Return the STDERR string
func (c *CMD) GetStderr() string {
	return c.stderr.String()
}

// Return the exit status integer
func (c *CMD) GetExitStatus() int {
	return c.exitStatus
}

// Check if the command is running
func (c *CMD) IsRunning() bool {
	return c.running
}

// Start the command in a goroutine
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

// Start the command and wait for it to finish running
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

// Stop the command by sending a SIGINT
func (c *CMD) Stop() error {
	if c.cmd == nil {
		return nil
	}

	return c.cmd.Process.Signal(os.Interrupt)
}

// Kill the command by sending a SIGKILL
func (c *CMD) Kill() error {
	if c.cmd == nil {
		return nil
	}

	return c.cmd.Process.Signal(os.Kill)
}