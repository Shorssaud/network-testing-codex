package codex

import (
	"bytes"
	"context"
	"fmt"

	"github.com/guseggert/clustertest/cluster"
	"github.com/guseggert/clustertest/cluster/basic"
)

var (
	Info  = Teal
	Warn  = Yellow
	Fatal = Red
)

var (
	Black   = Color("\033[1;30m%s\033[0m")
	Red     = Color("\033[1;31m%s\033[0m")
	Green   = Color("\033[1;32m%s\033[0m")
	Yellow  = Color("\033[1;33m%s\033[0m")
	Purple  = Color("\033[1;34m%s\033[0m")
	Magenta = Color("\033[1;35m%s\033[0m")
	Teal    = Color("\033[1;36m%s\033[0m")
	White   = Color("\033[1;37m%s\033[0m")
)

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

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
