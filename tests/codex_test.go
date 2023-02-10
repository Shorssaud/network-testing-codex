package codex

import (
	"bytes"
	"context"
	"fmt"
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
			nodes := c.MustNewNodes(2)

			group, groupCtx := errgroup.WithContext(context.Background())
			addrs := "127.0.0.1"
			for i, node := range nodes {
				node := node.Context(groupCtx)
				if i == 0 {
					group.Go(func() error {
						// host node
						stdout := &bytes.Buffer{}
						stderr := &bytes.Buffer{}
						// when running api on 8080, the host node will not be able to run due to nodeagent

						// gets node ip
						_, err := node.Run(cluster.StartProcRequest{
							Command: "hostname",
							Args:    []string{"-I"},
							Stdout:  stdout,
						})
						fmt.Println(stdout.String())

						_, err = node.StartProc(cluster.StartProcRequest{
							Command: "./build/codex",
							Args:    []string{"--metrics", "--api-port=8090", "--data-dir=`pwd`/Codex1", "--disc-port=8070", "--log-level=TRACE"},
							Stdout:  stdout,
							Stderr:  stderr,
						})
						if err != nil {
							t.Errorf(`starting host on node %d: %s`, i, err)
							t.Errorf("HOST EOutput: %s\n", stderr.String())
							return err
						}
						fmt.Println("---------------------")
						fmt.Printf("HOST Output: %s\n", stdout.String())

						// codex seems to exit upon function exit
						runout := &bytes.Buffer{}
						runerr := &bytes.Buffer{}
						// code below runs a local call to the api, WORKS only with localhost and not ip
						time.Sleep(2 * time.Second)
						_, err = node.Run(cluster.StartProcRequest{
							Command: "curl",
							Args:    []string{"http://" + addrs + ":8090/api/codex/v1/debug/info"},
							Stdout:  runout,
							Stderr:  runerr,
						})
						fmt.Println(runout)
						if err != nil {
							t.Errorf("%s", runerr)
						}
						// t.Logf("HOST Output: %s\n\n", stdout.String())
						// t.Logf("HOST EOutput: %s\n\n", stderr.String())

						return nil
					})
					time.Sleep(5 * time.Second)
					continue
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
			// group.Go(func() error {
			// 	// calls to the api local and with ip address, DOES NOT WORK
			// 	resp, err := http.Get("http://" + addrs + ":8090/api/codex/v1/debug/info")
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}
			// 	fmt.Println(resp)
			// 	defer resp.Body.Close()
			// 	return nil
			// })
			group.Wait()
		})

	}
	// run(t, "local cluster", local.MustNewCluster())
	run(t, "Docker cluster", docker.MustNewCluster().WithBaseImage("corbo12/nim-codex:v4"))
}
