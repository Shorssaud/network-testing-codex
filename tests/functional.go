package codex

import (
	"bytes"
	"context"

	"github.com/guseggert/clustertest/cluster"
	"github.com/guseggert/clustertest/cluster/basic"
)

func getIp(ctx context.Context, node *basic.Node) (string, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	proc, err := node.StartProc(cluster.StartProcRequest{
		Command: "hostname",
		Args:    []string{"-I"},
		Stdout:  stdout,
		Stderr:  stderr,
	})
	proc.Wait()
	if err != nil {
		return stderr.String(), err
	}
	return stdout.String(), nil
}

func createCodexInstance(ctx context.Context, node *basic.Node) (string, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	_, err := node.StartProc(cluster.StartProcRequest{
		Command: "./build/codex",
		Args:    []string{"--metrics", "--api-port=8090", "--data-dir=`pwd`/Codex1", "--disc-port=8070", "--log-level=TRACE"},
		Stdout:  stdout,
		Stderr:  stderr,
	})
	if err != nil {
		return stderr.String(), err
	}
	return stdout.String(), nil
}

func debugInfoCall(ctx context.Context, node *basic.Node, addrs string) (string, int, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	proc, err := node.StartProc(cluster.StartProcRequest{
		Command: "curl",
		Args:    []string{"http://" + addrs + ":8090/api/codex/v1/debug/info"},
		Stdout:  stdout,
		Stderr:  stderr,
	})
	if err != nil {
		return stderr.String(), 1, err
	}
	code, _ := proc.Wait()
	if err != nil {
		return stderr.String(), code.ExitCode, err
	}
	return stdout.String(), code.ExitCode, nil
}
