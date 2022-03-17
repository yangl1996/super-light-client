package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"sync"
	"os/exec"
	"math/rand"
)

type RemoteError struct {
	inner   error
	problem string
}

func (e RemoteError) Error() string {
	if e.inner != nil {
		return e.problem + ": " + e.inner.Error()
	} else {
		return e.problem
	}
}

func dispatchBwTest(args []string) {
	command := flag.NewFlagSet("exp", flag.ExitOnError)
	serverListFilePath := command.String("l", "servers.json", "path to the server list file")
	install := command.String("install", "", "install the given binary")
	generate := command.Int("generate", 0, "generate the dirty tree with the given size")
	dim := command.Int("dim", 50, "set the dimension of the tree")
	serve := command.Bool("serve", false, "serve the dirty tree")

	command.Parse(args[0:])

	if *serverListFilePath == "" {
		fmt.Println("missing server list")
		os.Exit(1)
	}

	// parse the server list
	servers := ReadServerInfo(*serverListFilePath)

	clients := make([]*ssh.Client, len(servers))
	connWg := &sync.WaitGroup{}	// wait for the ssh connection
	connWg.Add(len(servers))
	for i, s := range servers {
		go func(i int, s Server) {
			defer connWg.Done()
			client, err := connectSSH(s.User, s.PublicIP, s.Port, s.KeyPath)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("Connected to %v\n", s.Location)
			clients[i] = client
		}(i, s)
	}
	connWg.Wait()

	if *install != "" {
		fn := func(s Server, c *ssh.Client) error {
			return uploadFile(s, *install, "super-light-client")
		}
		runAll(servers, clients, fn)
	}

	if *generate != 0 {
		fn := func(s Server, c *ssh.Client) error {
			var diff int
			for diff == 0 {
				diff = rand.Intn(*generate)
			}
			if err := cleanUpLedger(c); err != nil {
				return err
			}
			sess, err := c.NewSession()
			if err != nil {
				return err
			}
			defer sess.Close()
			cmd := fmt.Sprintf("./super-light-client build -diff %d -dim %d -size %d", diff, *dim, *generate)
			return sess.Run(cmd)
		}
		runAll(servers, clients, fn)
	}

	if *serve {
		fn := func(s Server, c *ssh.Client) error {
			if err := killServer(c); err != nil {
				return err
			}
			sess, err := c.NewSession()
			if err != nil {
				return err
			}
			defer sess.Close()
			cmd := fmt.Sprintf("./super-light-client serve")
			return sess.Run(cmd)
		}
		runAll(servers, clients, fn)
	}
}


func runAll(servers []Server, clients []*ssh.Client, fn func(Server, *ssh.Client) error) error {
	if len(servers) != len(clients) {
		panic("incorrect")
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(clients))
	for i := range clients {
		go func(i int, s Server, c *ssh.Client) {
			defer wg.Done()
			err := fn(s, c)
			if err != nil {
				switch err := err.(type) {
				case *exec.ExitError:
					fmt.Printf("error executing local command for server %v: %s\n", i, err.Stderr)
				case *ssh.ExitError:
					fmt.Printf("error executing command on server %v: %s\n", i, err.Msg())
				default:
					fmt.Printf("error executing on server %v: %v\n", i, err)
				}
			}
		}(i, servers[i], clients[i])
	}
	wg.Wait()
	return nil
}

// TODO: use go-native ssh
func copyBackFile(s Server, from, dest string) error {
	fromStr := fmt.Sprintf("%s@%s:%s", s.User, s.PublicIP, from)
	cmdArgs := []string{"-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "-i", s.KeyPath, fromStr, dest}
	proc := exec.Command("scp", cmdArgs...)
	err := proc.Run()
	if err != nil {
		return err
	}
	return nil
}

func uploadFile(s Server, from, dest string) error {
	toStr := fmt.Sprintf("%s@%s:%s", s.User, s.PublicIP, dest)
	cmdArgs := []string{"-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "-i", s.KeyPath, from, toStr}
	//fmt.Println(cmdArgs)
	proc := exec.Command("scp", cmdArgs...)
	err := proc.Run()
	if err != nil {
		return err
	}
	return nil
}

func killServer(c *ssh.Client) error {
	pkill, err := c.NewSession()
	if err != nil {
		return RemoteError{err, "error creating session"}
	}
	pkill.Run(`killall -w super-light-client`)
	pkill.Close()
	return nil
}

func cleanUpLedger(c *ssh.Client) error {
	rmrf, err := c.NewSession()
	if err != nil {
		return RemoteError{err, "error creating session"}
	}
	rmrf.Run(`rm -rf tree.pogreb`)
	rmrf.Close()
	return nil
}

