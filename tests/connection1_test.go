package codex

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/guseggert/clustertest/cluster"
	"github.com/guseggert/clustertest/cluster/basic"
	"github.com/guseggert/clustertest/cluster/docker"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestDownload1(t *testing.T) {
	run := func(t *testing.T, name string, impl cluster.Cluster) {
		t.Run(name, func(t *testing.T) {
			var spr string
			var cid string
			var ip string
			t.Parallel()

			// Create the cluster.
			c := basic.New(impl)
			t.Cleanup(c.MustCleanup)

			t.Logf("Launching %s nodes", name)
			nodes := c.MustNewNodes(2)

			group, groupCtx := errgroup.WithContext(context.Background())
			addrs := "127.0.0.1"
			for i, node := range nodes {
				id := i
				node := node.Context(groupCtx)
				group.Go(func() error {
					if id == 0 {
						ip, err := getIp(groupCtx, node)
						if err != nil {
							t.Error(Fatal("failed to get ip: %s", err))
							t.Error(Fatal("HOST EOutput: %s\n", ip))
						}
						t.Log(Info("ip: " + ip))
						stdout := &bytes.Buffer{}
						stderr := &bytes.Buffer{}
						runout := &bytes.Buffer{}
						runerr := &bytes.Buffer{}
						_, err = node.StartProc(cluster.StartProcRequest{
							Command: "./build/codex",
							Args:    []string{"--metrics", "--api-port=8090", "--data-dir=`pwd`/Codex1", "--disc-port=8070", "--log-level=TRACE", "-e=" + ip[:len(ip)-2]},
							Stdout:  stdout,
							Stderr:  stderr,
						})

						if err != nil {
							t.Error(Fatal("HOST EOutput: %s\n", stderr.String()))
							return err
						}
						time.Sleep(2 * time.Second)

						out, _, err := debugInfoCall(groupCtx, node, addrs)
						if err != nil {
							t.Error(Fatal("failed to get debug info: %s", err))
						}
						t.Error(Fatal("NOPE", stderr.String()))
						temp := strings.Split(out, "\"spr\":")
						if len(temp) < 2 {
							t.Error(Fatal("failed to get spr: ", temp))
						}
						t.Log(Info("debug info: ", out))
						t.Log(Info("debug info: ", temp[1][1:len(temp[1])-1]))
						spr = temp[1][1 : len(temp[1])-2]
						node.SendFile("tests/dog1.txt", bytes.NewBuffer([]byte("hello my dog")))
						proc, err := node.StartProc(cluster.StartProcRequest{
							Command: "curl",
							Args:    []string{"-vvv", "-H", "content-type: application/octet-stream", "-H", "Expect:", "-T", "tests/dog1.txt", "http://" + addrs + ":8090/api/codex/v1/upload", "-X", "POST"},
							Stdout:  runout,
							Stderr:  runerr,
						})
						if err != nil {
							t.Error(Fatal("HOST EOutput: ", runout.String()))
							return err
						}
						_, err = proc.Wait()
						if err != nil {
							t.Error(Fatal("HOST EOutput: ", runout.String()))
							return err
						}
						t.Log(Info(runout.String()))
						t.Log(stderr.String())
						cid = runout.String()
					}
					if id == 1 {
						stdout := &bytes.Buffer{}
						stderr := &bytes.Buffer{}
						runout := &bytes.Buffer{}
						runerr := &bytes.Buffer{}
						_, err := node.StartProc(cluster.StartProcRequest{
							Command: "./build/codex",
							Args:    []string{"--metrics", "--api-port=8091", "--data-dir=`pwd`/Codex1", "--disc-port=8071", "--log-level=TRACE", "--bootstrap-node=" + spr},
							Stdout:  stdout,
							Stderr:  stderr,
						})
						time.Sleep(2 * time.Second)
						if err != nil {
							t.Error(Fatal("HOST EOutput: ", stderr.String()))
						}
						t.Log(stdout.String())
						t.Log(stderr.String())
						stdout.Reset()
						stderr.Reset()
						t.Log(Info("CID: " + cid))
						proc, err := node.StartProc(cluster.StartProcRequest{
							Command: "curl",
							Args:    []string{"-vvv", "http://" + ip + ":8090/api/codex/v1/download/" + cid, "--output", "tests/dog2.txt"},
							Stdout:  runout,
						})
						if err != nil {
							t.Error(Fatal("HOST EOutput: ", stderr.String()))
						}
						code, err := proc.Wait()
						if err != nil {
							t.Error(Fatal("HOST EOutput: ", stderr.String()))
						}
						t.Log(Info("download return code: ", code.ExitCode))
						t.Log(Info(runout.String()))
						t.Log(runerr.String())
						runout.Reset()
						runerr.Reset()
						_, err = node.Run(cluster.StartProcRequest{
							Command: "cat",
							Args:    []string{"tests/dog2.txt"},
							Stdout:  runout,
							Stderr:  runerr,
						})
						t.Log(Info(runout.String()))
						t.Log(runerr.String())
						assert.Equal(t, "hello my dog", runout.String())
					}
					return nil
				})
				time.Sleep(5 * time.Second)
			}
			group.Wait()
		})

	}
	// run(t, "local cluster", local.MustNewCluster())
	run(t, "Docker cluster", docker.MustNewCluster().WithBaseImage("corbo12/nim-codex:v4"))
}
