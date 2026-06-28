package infra

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/james-barrow/golang-ipc"

	"github.com/egovelox/mozeidon/browser/core/models"
	"github.com/egovelox/mozeidon/debug"
)

type IpcClient struct {
	*ipc.Client
}

type EndMessage struct {
	End string `json:"data"`
}

// printErr emits an error to stderr using the same {"error": "..."} envelope as
// core.PrintError (which this low-level package can't import without a cycle).
func printErr(message string) {
	b, _ := json.Marshal(map[string]string{"error": message})
	fmt.Fprintln(os.Stderr, string(b))
}

func (ipc *IpcClient) Send(
	cmd models.Command,
) <-chan models.CommandResult {

	debug.Logf("sending command %q (args=%q)", cmd.Command, cmd.Args)
	jsonCmd, _ := json.Marshal(cmd)
	ipc.Write(1, []byte(jsonCmd))

	channel := make(chan models.CommandResult)
	go func() {
		defer close(channel)
		for {
			message, err := ipc.Read()
			if err != nil {
				debug.Logf("ipc read error: %v", err)
				printErr("[Error] Lost connection to the Mozeidon-Z native host while reading the response. The browser or extension may have disconnected — reopen Firefox with the Mozeidon-Z extension and retry.")
				os.Exit(1)
			}
			if message.MsgType > 0 {
				if string(message.Data) == `{"data":"end"}` {
					debug.Logf("received end-of-stream; closing")
					break
				}
				channel <- models.CommandResult{Data: message.Data}
			}
		}
	}()
	return channel
}

func NewIpcClient(host string) *IpcClient {
	config := ipc.ClientConfig{
		Encryption: true,
		Timeout:    2,
		RetryTimer: 0,
	}

	debug.Logf("dialing ipc %s", host)
	ipc, err := ipc.StartClient(host, &config)
	if err != nil {
		debug.Logf("ipc StartClient error: %v", err)
		printErr(fmt.Sprintf("[Error] Could not reach the Mozeidon-Z native host (ipc: %s). Is a browser open with the Mozeidon-Z extension connected? Run 'mozeidon-z profiles get' to list connected browsers.", host))
		os.Exit(1)
	}

	for {
		message, err := ipc.Read()
		if err != nil {
			debug.Logf("ipc handshake read error: %v", err)
			printErr(fmt.Sprintf("[Error] Lost connection to the Mozeidon-Z native host (ipc: %s) during handshake. Reopen the browser/extension and retry.", host))
			os.Exit(1)
		}
		if message.MsgType == -1 && message.Status == "Connected" {
			debug.Logf("connected to ipc %s", host)
			break
		}
	}
	return &IpcClient{ipc}
}
