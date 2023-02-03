package codex

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/guseggert/clustertest/cluster"
	"github.com/guseggert/clustertest/cluster/basic"
	"github.com/guseggert/clustertest/cluster/docker"
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
			nodes := c.MustNewNodes(5)

			group, groupCtx := errgroup.WithContext(context.Background())
			addrs := "/ip4/0.0.0.0/tcp/50880"
			for i, node := range nodes {
				node := node.Context(groupCtx)
				if i == 0 {
					// host node
					stdout := &bytes.Buffer{}
					stderr := &bytes.Buffer{}
					_, err := node.StartProc(cluster.StartProcRequest{
						Command: "./build/codex",
						Args:    []string{"--nat=100.80.179.146", "--metrics", "--api-port=8080", "--data-dir=`pwd`/Codex1", "--disc-port=8090", "--log-level=TRACE", "-i=" + addrs, "-q=1099511627776"},
						Stdout:  stdout,
						Stderr:  stderr,
					})
					if err != nil {
						t.Errorf(`starting client on node %d: %s`, i, err)
						continue
					}
					// extract spr from host node output
					time.Sleep(2 * time.Second)
					if err != nil {
						t.Errorf(`waiting for client on node %d: %s`, i, err)
						continue
					}
					t.Logf("HOST Output: %s\n\n", stdout.String())
					t.Logf("HOST EOutput: %s\n\n", stderr.String())
					// t.Log("HOST proc: ", code)

					continue
				} else {
					group.Go(func() error {
						// wait for host to start
						resp, err := http.Get("http://127.0.0.1:8080/api/codex/v1/info")
						if err != nil {
							log.Fatal(err)
						}
						fmt.Println(resp)
						defer resp.Body.Close()
						return nil
					})
				} // else {
				// 	group.Go(func() error {
				// 		stdout := &bytes.Buffer{}
				// 		stderr := &bytes.Buffer{}
				// 		_, err := node.StartProc(cluster.StartProcRequest{
				// 			Command: "./build/codex",
				// 			Args:    []string{"--data-dir=\"$(pwd)/Codex2\"", "--api-port=8081", "--disc-port=8091", "--bootstrap-node=$(curl  http://127.0.0.1:8080/api/codex/v1/debug/info 2>/dev/null | jq -r .spr)", "--log-level=TRACE", "-i=/ip4/0.0.0.0/tcp/50881", "-q=1099511627776"},
				// 			Stdout:  stdout,
				// 			Stderr:  stderr,
				// 		})
				// 		if err != nil {
				// 			t.Errorf(`starting client on node %d: %s`, i, err)
				// 			return nil
				// 		}
				// 		t.Logf("CLIENT Output: %s\n\n", stdout.String())
				// 		t.Logf("CLIENT EOutput: %s\n\n", stderr.String())
				// 		// t.Log("CLIENT proc: ", code)

				// 		return nil
				// 	})
				// }
			}
			group.Wait()
			// TODO ping host using get request
		})
	}

	// run(t, "local cluster", local.MustNewCluster())
	run(t, "Docker cluster", docker.MustNewCluster().WithBaseImage("corbo12/nim-codex:v3"))
}
