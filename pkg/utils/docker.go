package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pkg/errors"

	"github.com/oam-dev/velad/pkg/apis"
)

// ExecResult is the result of Docker Exec
type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

// Exec create an exec command and return its ID
func Exec(ctx context.Context, cli *client.Client, containerID string, command []string) (types.IDResponse, error) {
	config := types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          command,
	}

	return cli.ContainerExecCreate(ctx, containerID, config)
}

// InspectExecResp helps turns types.IDResponse into a ExecResult
func InspectExecResp(ctx context.Context, cli *client.Client, id string) (ExecResult, error) {
	var execResult ExecResult

	resp, err := cli.ContainerExecAttach(ctx, id, types.ExecStartCheck{})
	if err != nil {
		return execResult, err
	}
	defer resp.Close()

	// read the output
	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return execResult, err
		}
		break

	case <-ctx.Done():
		return execResult, ctx.Err()
	}

	stdout, err := io.ReadAll(&outBuf)
	if err != nil {
		return execResult, err
	}
	stderr, err := io.ReadAll(&errBuf)
	if err != nil {
		return execResult, err
	}
	res, err := cli.ContainerExecInspect(ctx, id)
	if err != nil {
		return execResult, err
	}

	execResult.ExitCode = res.ExitCode
	execResult.StdOut = string(stdout)
	execResult.StdErr = string(stderr)
	return execResult, nil
}

// GetTokenFromCluster returns the token for k3d cluster
func GetTokenFromCluster(ctx context.Context, clusterName string) (string, error) {
	dockerCli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", errors.Wrap(err, "failed to create docker client")
	}
	defer func(dockerCli *client.Client) {
		_ = dockerCli.Close()
	}(dockerCli)

	containers, err := dockerCli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return "", errors.Wrap(err, "failed to list containers")
	}
	var ID string
	for _, c := range containers {
		for _, name := range c.Names {
			if name == fmt.Sprintf("/k3d-velad-cluster-%s-server-0", clusterName) {
				ID = c.ID
			}
		}
	}
	if ID == "" {
		return "", errors.Errorf("no cluster with name %s found.", clusterName)
	}
	exec, err := Exec(ctx, dockerCli, ID, []string{"cat", apis.K3sTokenPath})
	if err != nil {
		return "", errors.Wrap(err, "failed to create docker exec command")
	}
	resp, err := InspectExecResp(ctx, dockerCli, exec.ID)
	if err != nil {
		return "", errors.Wrap(err, "failed to inspect exec command result")
	}
	if resp.ExitCode != 0 {
		return "", errors.Errorf("failed to get token, exit code: %d, stderr: %s", resp.ExitCode, resp.StdErr)
	}
	if resp.StdOut == "" {
		return "", errors.Errorf("token is empty, stderr: %s", resp.StdErr)
	}
	return resp.StdOut, nil
}
