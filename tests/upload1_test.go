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

func TestUpload1(t *testing.T) {
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
				node := node.Context(groupCtx)
				if i <= 1 {
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
						runout := &bytes.Buffer{}
						runerr := &bytes.Buffer{}
						time.Sleep(2 * time.Second)
						if i == 1 {
							node.SendFile("tests/dog1.txt", bytes.NewBuffer([]byte("hello my dog")))
							for x := 0; x < 2; x++ {
								group.Go(func() error {
									proc, err := node.StartProc(cluster.StartProcRequest{
										Command: "curl",
										Args:    []string{"-vvv", "-H", "\"content-type: application/octet-stream\"", "-H", "Expect:", "-T", "tests/dog1.txt", "http://" + addrs + ":8090/api/codex/v1/upload", "-X", "POST"},
										Stdout:  runout,
									})
									if err != nil {
										t.Error(Fatal("HOST EOutput: %s\n", err))
										return err
									}
									code, err := proc.Wait()
									if err != nil {
										t.Errorf(Fatal("HOST EOutput: %s\n", err))
										return err
									}
									t.Log(Info("HOST Output: ") + runout.String())
									assert.Equal(t, 0, code.ExitCode, "HOST EOutput: %s\n", runerr.String())
									return nil
								})
							}
						}
						return nil
					})
				}
			}
			group.Wait()
		})

	}
	// run(t, "local cluster", local.MustNewCluster())
	run(t, "Docker cluster", docker.MustNewCluster().WithBaseImage("corbo12/nim-codex:v4"))
}
