// Copyright 2018 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/u-root/u-root/pkg/bb"
	"github.com/u-root/u-root/pkg/uroot/util"
)

func run() {
	name := filepath.Base(os.Args[0])
	if err := bb.Run(name); err != nil {
		log.Fatalf("%s: %v", name, err)
	}
}

func main() {
	os.Args[0] = util.ResolveUntilLastSymlink(os.Args[0])

	run()
}

func init() {
	m := func() {
		// Use argv[1] as the name.
		os.Args = os.Args[1:]
		run()
	}
	bb.Register("bb", bb.Noop, m)
	bb.RegisterDefault(bb.Noop, m)
}