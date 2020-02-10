// Copyright 2017 Groove ID, Inc.
//
// All information contained herein is the property of Groove ID, Inc. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Groove ID, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Groove ID, Inc.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/nightlyone/lockfile"
)

func main() {
	var path string
	flag.StringVar(&path, "p", "", "Path to the lock file")
	flag.StringVar(&path, "path", "", "Path to the lock file")
	flag.Parse()
	if len(flag.Args()) == 0 || path == "" {
		fmt.Printf("usage: %s -p <path> -- <command> [args...]\n", os.Args[0])
		os.Exit(1)
	}

	lock, err := lockfile.New(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot initialize lock file: %v\n", err)
		os.Exit(1)
	}

	for count := 0; true; count++ {
		err = lock.TryLock()
		if err == nil {
			if count > 0 {
				fmt.Fprintf(os.Stderr, "%s: got lock\n", path)
			}
			break
		} else if err == lockfile.ErrBusy {
			if count == 0 {
				fmt.Fprintf(os.Stderr, "%s: waiting for lock\n", path)
			}
			time.Sleep(100 * time.Millisecond)
		} else {
			fmt.Fprintf(os.Stderr, "cannot obtain lock: %v\n", err)
			os.Exit(1)
		}
	}

	cmd := exec.Command(flag.Args()[0], flag.Args()[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()

	lock.Unlock()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "cannot execute process: %v\n", err)
			os.Exit(1)
		}
	}
}
