// +build functional

//
// Copyright (c) 2019 The heketi Authors
//
// This file is licensed to you under your choice of the GNU Lesser
// General Public License, version 3 or any later version (LGPLv3 or
// later), as published by the Free Software Foundation,
// or under the Apache License, Version 2.0 <LICENSE-APACHE2 or
// http://www.apache.org/licenses/LICENSE-2.0>.
//
// You may not use this file except in compliance with those terms.
//

package tests

import (
	"os"
	"path"
	"testing"

	"github.com/heketi/tests"

	"github.com/heketi/heketi/pkg/glusterfs/api"
	"github.com/heketi/heketi/pkg/testutils"
	"github.com/heketi/heketi/server/config"
)

func TestBlockVolumeAllocDefaults(t *testing.T) {
	heketiServer := testutils.NewServerCtlFromEnv("..")
	origConf := path.Join(heketiServer.ServerDir, heketiServer.ConfPath)

	heketiServer.ConfPath = tests.Tempfile()
	defer os.Remove(heketiServer.ConfPath)
	CopyFile(origConf, heketiServer.ConfPath)

	defer func() {
		CopyFile(origConf, heketiServer.ConfPath)
		testutils.ServerRestarted(t, heketiServer)
		testCluster.Teardown(t)
		testutils.ServerStopped(t, heketiServer)
	}()

	testutils.ServerStarted(t, heketiServer)
	heketiServer.KeepDB = true
	testCluster.Setup(t, 3, 3)

	blockReq := &api.BlockVolumeCreateRequest{}
	blockReq.Size = 3
	blockReq.Hacount = 3

	// create a volume (and BHV) with default unset
	_, err := heketi.BlockVolumeCreate(blockReq)
	tests.Assert(t, err == nil, "expected err == nil, got:", err)

	t.Run("AllocFull", func(t *testing.T) {
		// explicitly set the default to "full"
		UpdateConfig(origConf, heketiServer.ConfPath, func(c *config.Config) {
			c.GlusterFS.SshConfig.BlockVolumePrealloc = "full"
		})
		testutils.ServerRestarted(t, heketiServer)

		_, err := heketi.BlockVolumeCreate(blockReq)
		tests.Assert(t, err == nil, "expected err == nil, got:", err)
	})
	t.Run("AllocNo", func(t *testing.T) {
		// explicitly set the default to "full"
		UpdateConfig(origConf, heketiServer.ConfPath, func(c *config.Config) {
			c.GlusterFS.SshConfig.BlockVolumePrealloc = "no"
		})
		testutils.ServerRestarted(t, heketiServer)

		_, err := heketi.BlockVolumeCreate(blockReq)
		tests.Assert(t, err == nil, "expected err == nil, got:", err)
	})
	t.Run("AllocInvalid", func(t *testing.T) {
		UpdateConfig(origConf, heketiServer.ConfPath, func(c *config.Config) {
			c.GlusterFS.SshConfig.BlockVolumePrealloc = "XXXfoobarXXX"
		})
		testutils.ServerRestarted(t, heketiServer)

		_, err := heketi.BlockVolumeCreate(blockReq)
		tests.Assert(t, err != nil, "expected err != nil, got:", err)

		// assert that no pending ops remain
		l, err := heketi.PendingOperationList()
		tests.Assert(t, err == nil, "expected err == nil, got:", err)
		tests.Assert(t, len(l.PendingOperations) == 0,
			"expected len(l.PendingOperations) == 0, got:", len(l.PendingOperations))
	})
}
