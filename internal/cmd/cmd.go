package cmd

import (
	"errors"
	"io"
	"log"
	"os/exec"
)

type Command struct {
	workingDir string
}

func NewCommand(workingDir string) *Command {
	return &Command{
		workingDir: workingDir,
	}
}

func (c *Command) Run(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	if c.workingDir != "" {
		cmd.Dir = c.workingDir
	}

	stderrpipe, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("cmd.StderrPipe: %v", err)
	}

	stdoutpipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("cmd.StdoutPipe: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		log.Printf("cmd.Start: %v", err)
	}

	stderr, err := io.ReadAll(stderrpipe)
	if err != nil {
		log.Printf("ReadAll: %v", err)
	}

	stdout, err := io.ReadAll(stdoutpipe)
	if err != nil {
		log.Printf("ReadAll: %v", err)
	}

	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			log.Printf("exit status %d", exiterr.ExitCode())
			if len(stderr) == 0 {
				return string(stdout), err
			}
		} else {
			log.Printf("cmd.Wait: %v", err)
		}
	}

	if err == nil {
		return string(stdout), nil
	}

	if len(stderr) > 0 {
		return string(stdout), errors.New(string(stderr))
	}

	return string(stdout), nil
}