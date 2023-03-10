package codex

import (
	"bytes"
	"context"
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
					// gets node ip
					ip, err := getIp(groupCtx, node)
					if err != nil {
						t.Error(Fatal("failed to get ip: %s", err))
						t.Error(Fatal("HOST EOutput: %s\n", ip))
					}
					t.Log(Info("ip: " + ip))

					stdout := &bytes.Buffer{}
					stderr := &bytes.Buffer{}

					_, err = node.StartProc(cluster.StartProcRequest{
						Command: "./build/codex",
						Args:    []string{"--metrics", "--api-port=8090", "--data-dir=`pwd`/Codex1", "--disc-port=8070", "--log-level=TRACE"},
						Stdout:  stdout,
					})
					if err != nil {
						t.Error(Fatal("HOST EOutput: %s\n", stderr.String()))
						return err
					}
					time.Sleep(2 * time.Second)

					if id == 1 {
						runout := &bytes.Buffer{}
						runerr := &bytes.Buffer{}
						_, err = node.Run(cluster.StartProcRequest{
							Command: "ls",
							Args:    []string{"tests"},
							Stdout:  runout,
							Stderr:  runerr,
						})
						t.Log(Info(runout.String()))
						node.SendFile("tests/dog1.txt", bytes.NewBuffer([]byte("hello my dog")))
						runout.Reset()
						runerr.Reset()
						for x := 0; x < 1; x++ {
							group.Go(func() error {
								runout.Reset()
								runerr.Reset()
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
								t.Log(runerr.String())
								cid := runout.String()
								runout.Reset()
								runerr.Reset()
								_, err = node.Run(cluster.StartProcRequest{
									Command: "ls",
									Args:    []string{"tests"},
									Stdout:  runout,
									Stderr:  runerr,
								})
								t.Log(Info(runout.String()))
								runout.Reset()
								runerr.Reset()
								t.Log(Info("CID: " + cid))
								proc, err = node.StartProc(cluster.StartProcRequest{
									Command: "curl",
									Args:    []string{"-vvv", "http://" + addrs + ":8090/api/codex/v1/download/" + cid, "--output", "tests/dog2.txt"},
									Stdout:  runout,
								})
								_, _ = proc.Wait()
								t.Log(Info(runout.String()))
								t.Log(runerr.String())
								runout.Reset()
								runerr.Reset()
								_, err = node.Run(cluster.StartProcRequest{
									Command: "ls",
									Args:    []string{"-l", "tests"},
									Stdout:  runout,
									Stderr:  runerr,
								})
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

								return nil
							})
						}
					}
					return nil
				})
			}
			group.Wait()
		})

	}
	// run(t, "local cluster", local.MustNewCluster())
	run(t, "Docker cluster", docker.MustNewCluster().WithBaseImage("corbo12/nim-codex:v4"))
}

//TODO download file with other node
