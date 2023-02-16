package codex

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/guseggert/clustertest/cluster"
	"github.com/guseggert/clustertest/cluster/basic"
	"github.com/guseggert/clustertest/cluster/docker"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestHello(t *testing.T) {
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
						// host node
						ip, err := getIp(groupCtx, node)
						if err != nil {
							t.Errorf("failed to get ip: %s", err)
							t.Errorf("HOST EOutput: %s\n", ip)
						}
						fmt.Println("ip: ", ip)
						output, err := createCodexInstance(groupCtx, node)
						if err != nil {
							t.Errorf(`starting host on node %d: %s`, i, err)
							t.Errorf("HOST EOutput: %s\n", output)
							return err
						}
						// code below runs a local call to the api, WORKS only with localhost and not ip
						time.Sleep(2 * time.Second)
						t.Log(output)
						for x := 0; x < 2; x++ {
							group.Go(func() error {
								output, code, err := debugInfoCall(groupCtx, node, addrs)
								if err != nil {
									t.Errorf("failed to get debug info: %s", err)
									t.Errorf("HOST EOutput: %s\n", output)
									return err
								}
								fmt.Println("---------------------")
								t.Log(output)
								assert.Equal(t, 0, code, "should be 200")
								t.Logf("HOST %d Exit code: %d\n", i, code)
								// fmt.Println(runout)
								return nil
							})
						}
						// t.Logf("HOST Output: %s\n\n", stdout.String())
						// t.Logf("HOST EOutput: %s\n\n", stderr.String())

						return nil
					})
					time.Sleep(5 * time.Second)
					continue
				}
			}
			group.Wait()
		})

	}
	// run(t, "local cluster", local.MustNewCluster())
	run(t, "Docker cluster", docker.MustNewCluster().WithBaseImage("corbo12/nim-codex:v4"))
}
