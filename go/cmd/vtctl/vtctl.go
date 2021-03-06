// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	log "github.com/golang/glog"
	"github.com/youtube/vitess/go/exit"
	"github.com/youtube/vitess/go/vt/logutil"
	myproto "github.com/youtube/vitess/go/vt/mysqlctl/proto"
	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/vtctl"
	"github.com/youtube/vitess/go/vt/wrangler"
)

var (
	waitTime        = flag.Duration("wait-time", 24*time.Hour, "time to wait on an action")
	lockWaitTimeout = flag.Duration("lock-wait-timeout", time.Minute, "time to wait for a lock before starting an action")
)

func init() {
	logger := logutil.NewConsoleLogger()
	flag.CommandLine.SetOutput(logutil.NewLoggerWriter(logger))
	flag.Usage = func() {
		logger.Printf("Usage: %s [global parameters] command [command parameters]\n", os.Args[0])
		logger.Printf("\nThe global optional parameters are:\n")
		flag.PrintDefaults()
		logger.Printf("\nThe commands are listed below, sorted by group. Use '%s <command> -h' for more help.\n\n", os.Args[0])
		vtctl.PrintAllCommands(logger)
	}
}

// signal handling, centralized here
func installSignalHandlers(wr *wrangler.Wrangler) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan
		// we got a signal, cancel the current wrangler context
		wr.Cancel()
	}()
}

func main() {
	defer exit.RecoverAll()
	defer logutil.Flush()

	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		exit.Return(1)
	}
	action := args[0]

	startMsg := fmt.Sprintf("USER=%v SUDO_USER=%v %v", os.Getenv("USER"), os.Getenv("SUDO_USER"), strings.Join(os.Args, " "))

	if syslogger, err := syslog.New(syslog.LOG_INFO, "vtctl "); err == nil {
		syslogger.Info(startMsg)
	} else {
		log.Warningf("cannot connect to syslog: %v", err)
	}

	topoServer := topo.GetServer()
	defer topo.CloseServers()

	wr := wrangler.New(logutil.NewConsoleLogger(), topoServer, *waitTime, *lockWaitTimeout)
	installSignalHandlers(wr)

	err := vtctl.RunCommand(wr, args)
	switch err {
	case vtctl.ErrUnknownCommand:
		flag.Usage()
		exit.Return(1)
	case nil:
		// keep going
	default:
		log.Errorf("action failed: %v %v", action, err)
		exit.Return(255)
	}
}

type rTablet struct {
	*topo.TabletInfo
	*myproto.ReplicationStatus
}

type rTablets []*rTablet

func (rts rTablets) Len() int { return len(rts) }

func (rts rTablets) Swap(i, j int) { rts[i], rts[j] = rts[j], rts[i] }

// Sort for tablet replication.
// master first, then i/o position, then sql position
func (rts rTablets) Less(i, j int) bool {
	// NOTE: Swap order of unpack to reverse sort
	l, r := rts[j], rts[i]
	// l or r ReplicationPosition would be nil if we failed to get
	// the position (put them at the beginning of the list)
	if l.ReplicationStatus == nil {
		return r.ReplicationStatus != nil
	}
	if r.ReplicationStatus == nil {
		return false
	}
	var lTypeMaster, rTypeMaster int
	if l.Type == topo.TYPE_MASTER {
		lTypeMaster = 1
	}
	if r.Type == topo.TYPE_MASTER {
		rTypeMaster = 1
	}
	if lTypeMaster < rTypeMaster {
		return true
	}
	if lTypeMaster == rTypeMaster {
		return !l.Position.AtLeast(r.Position)
	}
	return false
}

func sortReplicatingTablets(tablets []*topo.TabletInfo, stats []*myproto.ReplicationStatus) []*rTablet {
	rtablets := make([]*rTablet, len(tablets))
	for i, status := range stats {
		rtablets[i] = &rTablet{tablets[i], status}
	}
	sort.Sort(rTablets(rtablets))
	return rtablets
}
