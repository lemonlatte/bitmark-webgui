// Copyright (c) 2014-2015 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file

package services

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/bitmark-inc/bitmark-webgui/configuration"
	"github.com/bitmark-inc/bitmark-webgui/fault"
	"github.com/bitmark-inc/bitmark-webgui/structs"
	"github.com/bitmark-inc/bitmark-webgui/utils"
	"github.com/bitmark-inc/bitmarkd/payment/bitcoin"
	"github.com/bitmark-inc/bitmarkd/rpc"
	"github.com/bitmark-inc/logger"
	"net"
	netrpc "net/rpc"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Bitmarkd struct {
	sync.RWMutex
	initialised bool
	log         *logger.L
	configFile  string
	process     *os.Process
	running     bool
	ModeStart   chan bool
}

func (bitmarkd *Bitmarkd) Initialise(configFile string) error {
	bitmarkd.Lock()
	defer bitmarkd.Unlock()

	if bitmarkd.initialised {
		return fault.ErrAlreadyInitialised
	}

	bitmarkd.configFile = configFile

	bitmarkd.log = logger.New("service-bitmarkd")
	if nil == bitmarkd.log {
		return fault.ErrInvalidLoggerChannel
	}

	bitmarkd.running = false
	bitmarkd.ModeStart = make(chan bool, 1)

	// all data initialised
	bitmarkd.initialised = true
	return nil
}

func (bitmarkd *Bitmarkd) Finalise() error {
	bitmarkd.Lock()
	defer bitmarkd.Unlock()

	if !bitmarkd.initialised {
		return fault.ErrNotInitialised
	}

	bitmarkd.initialised = false
	return nil
}

func (bitmarkd *Bitmarkd) IsRunning() bool {
	return bitmarkd.running
}

func (bitmarkd *Bitmarkd) Setup(chain string, webguiConfigFile string, webguiConfig *configuration.Configuration) error {
	bitmarkConfigFile := filepath.Join(webguiConfig.DataDirectory, fmt.Sprintf("bitmarkd-%s", chain), "bitmarkd.conf")

	if bitmarkd.running {
		bitmarkd.ModeStart <- false
	}

	bitmarkd.configFile = bitmarkConfigFile

	webguiConfig.BitmarkChain = chain
	webguiConfig.BitmarkConfigFile = bitmarkConfigFile

	err := EnsureFile(bitmarkConfigFile)
	if err != nil {
		return err
	}

	if bitmarkConfigs, err := structs.NewBitmarkdConfiguration(bitmarkConfigFile); nil != err {
		return err
	} else {
		bitmarkConfigs.Chain = chain
		switch chain {
		case structs.Bitmark:
			bitmarkConfigs.Bitcoin = bitcoin.Configuration{
				URL: "https://spoon.live.bitmark.com:17011/rpc-call",
			}
		case structs.Testing:
			bitmarkConfigs.Bitcoin = bitcoin.Configuration{
				URL: "https://spoon.test.bitmark.com:17555/rpc-call",
			}
		}
		bitmarkConfigs.SaveToJson(bitmarkConfigFile)
	}

	cmdResult, err := SimpleCmd("bitmarkd", "--config-file="+bitmarkConfigFile, "gen-peer-identity")
	if err != nil {
		if !strings.Contains(cmdResult, "file already exists") {
			return err
		}
		bitmarkd.log.Warn(cmdResult)
	}

	cmdResult, err = SimpleCmd("bitmarkd", "--config-file="+bitmarkConfigFile, "gen-proof-identity")
	if err != nil {
		if !strings.Contains(cmdResult, "file already exists") {
			return err
		}
		bitmarkd.log.Warn(cmdResult)
	}

	cmdResult, err = SimpleCmd("bitmarkd", "--config-file="+bitmarkConfigFile, "gen-rpc-cert")
	if err != nil {
		if !strings.Contains(cmdResult, "file already exists") {
			return err
		}
		bitmarkd.log.Warn(cmdResult)
	}

	return configuration.UpdateConfiguration(webguiConfigFile, webguiConfig)
}

func (bitmarkd *Bitmarkd) Run(args interface{}, shutdown <-chan struct{}) {
loop:
	for {
		select {

		case <-shutdown:
			break loop
		case start := <-bitmarkd.ModeStart:
			if start {
				bitmarkd.startBitmarkd()
			} else {
				bitmarkd.stopBitmarkd()
			}
		}

	}
	close(bitmarkd.ModeStart)
}

func (bitmarkd *Bitmarkd) startBitmarkd() error {
	if bitmarkd.running {
		bitmarkd.log.Errorf("Start bitmarkd failed: %v", fault.ErrBitmarkdIsRunning)
		return fault.ErrBitmarkdIsRunning
	}

	// Check bitmarkConfigFile exists
	bitmarkd.log.Infof("bitmark config file: %s\n", bitmarkd.configFile)
	if !utils.EnsureFileExists(bitmarkd.configFile) {
		bitmarkd.log.Errorf("Start bitmarkd failed: %v", fault.ErrNotFoundConfigFile)
		return fault.ErrNotFoundConfigFile
	}

	bitmarkd.running = true
	stopped := make(chan bool, 1)

	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-stopped:
			return
		case <-ch:
			bitmarkd.stopBitmarkd()
		}
	}()

	go func() {
		defer func() {
			stopped <- true
		}()
		for bitmarkd.running {
			// start bitmarkd as sub process
			cmd := exec.Command("bitmarkd", "--config-file="+bitmarkd.configFile)
			// start bitmarkd as sub process
			stderr, err := cmd.StderrPipe()
			if err != nil {
				bitmarkd.log.Errorf("Error: %v", err)
				continue
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				bitmarkd.log.Errorf("Error: %v", err)
				continue
			}
			if err := cmd.Start(); nil != err {
				continue
			}

			bitmarkd.process = cmd.Process
			bitmarkd.log.Infof("process id: %d", cmd.Process.Pid)
			stdeReader := bufio.NewReader(stderr)
			stdoReader := bufio.NewReader(stdout)

			go func() {
				for {
					stde, err := stdeReader.ReadString('\n')
					bitmarkd.log.Errorf("bitmarkd stderr: %q", stde)
					if nil != err {
						bitmarkd.log.Errorf("Error: %v", err)
						return
					}
				}
			}()

			go func() {
				for {
					stdo, err := stdoReader.ReadString('\n')
					bitmarkd.log.Infof("bitmarkd stdout: %q", stdo)
					if nil != err {
						bitmarkd.log.Errorf("Error: %v", err)
						return
					}
				}
			}()

			if err := cmd.Wait(); nil != err {
				if bitmarkd.running {
					bitmarkd.log.Errorf("bitmarkd has terminated unexpectedly. failed: %v", err)
					bitmarkd.log.Errorf("bitmarkd will be restarted in 1 second...")
					time.Sleep(time.Second)
				}
				bitmarkd.process = nil
			}
		}
	}()

	// wait for 1 second if cmd has no error then return nil
	time.Sleep(time.Second * 1)
	return nil

}

func (bitmarkd *Bitmarkd) stopBitmarkd() error {
	if !bitmarkd.running {
		bitmarkd.log.Errorf("Stop bitmarkd failed: %v", fault.ErrBitmarkdIsNotRunning)
		return fault.ErrBitmarkdIsNotRunning
	}
	bitmarkd.running = false

	// if err := bitmarkd.process.Signal(os.Interrupt); nil != err {
	// bitmarkd.log.Errorf("Send interrupt to bitmarkd failed: %v", err)
	if bitmarkd.process == nil {
		return nil
	}
	if err := bitmarkd.process.Signal(os.Kill); nil != err {
		bitmarkd.log.Errorf("Send kill to bitmarkd failed: %v", err)
		return err
	}

	bitmarkd.log.Infof("Stop bitmarkd. PID: %d", bitmarkd.process.Pid)
	bitmarkd.process = nil
	return nil
}

func (bitmarkd *Bitmarkd) GetInfo(client *netrpc.Client) (*rpc.InfoReply, error) {

	var reply rpc.InfoReply
	if err := client.Call("Node.Info", rpc.InfoArguments{}, &reply); err != nil {
		bitmarkd.log.Errorf("Node.Info error: %v\n", err)
		return nil, fault.ErrNodeInfoRequestFail
	}

	return &reply, nil
}

// connect to bitmarkd RPC
func (bitmarkd *Bitmarkd) Connect(connect string) (net.Conn, error) {

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", connect, tlsConfig)
	if nil != err {
		return nil, err
	}

	return conn, nil
}
