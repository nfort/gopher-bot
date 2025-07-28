package cmd

import (
	"context"
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

func (c *Command) Run(ctx context.Context, name string, arg ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, arg...)
	if c.workingDir != "" {
		cmd.Dir = c.workingDir
	}

	stderrpipe, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("cmd.StderrPipe: %v", err)
		return "", err
	}

	stdoutpipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("cmd.StdoutPipe: %v", err)
		return "", nil
	}

	err = cmd.Start()
	if err != nil {
		log.Printf("cmd.Start: %v", err)
		return "", nil
	}

	stderr, err := io.ReadAll(stderrpipe)
	if err != nil {
		log.Printf("ReadAll: %v", err)
		return "", nil
	}

	stdout, err := io.ReadAll(stdoutpipe)
	if err != nil {
		log.Printf("ReadAll: %v", err)
		return "", nil
	}

	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			log.Printf("exit status %d", exiterr.ExitCode())
			log.Printf("exiterr: %s", exiterr.Error())
			log.Printf("stderr: %s", string(stderr))
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
