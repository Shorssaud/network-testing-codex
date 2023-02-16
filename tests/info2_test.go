// This test is a counter test to check if the tests are up to date and running correctly by applying an incorrect ip

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

func TestInfo2(t *testing.T) {
	run := func(t *testing.T, name string, impl cluster.Cluster) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c := basic.New(impl)
			t.Cleanup(c.MustCleanup)

			t.Logf("Launching %s nodes", name)
			nodes := c.MustNewNodes(2)

			group, groupCtx := errgroup.WithContext(context.Background())
			addrs := "127.0.0.2" // incorrect ip
			for i, node := range nodes {
				node := node.Context(groupCtx)
				if i <= 1 {
					group.Go(func() error {
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
								assert.NotEqual(t, 0, code, "should be not be 0")
								t.Logf("HOST %d Exit code: %d\n", i, code)
								return nil
							})
						}
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
