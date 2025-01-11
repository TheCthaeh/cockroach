// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package nodelocal

import (
	"testing"

	"github.com/cockroachdb/cockroach/pkg/cloud/cloudtestutils"
	"github.com/cockroachdb/cockroach/pkg/security/username"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
)

func TestPutLocal(t *testing.T) {
	defer leaktest.AfterTest(t)()

	p, cleanupFn := testutils.TempDir(t)
	defer cleanupFn()

	testSettings := cluster.MakeTestingClusterSettings()
	testSettings.ExternalIODir = p
	dest := MakeLocalStorageURI(p)

	info := cloudtestutils.StoreInfo{
		URI:           dest,
		User:          username.RootUserName(),
		ExternalIODir: p,
	}
	cloudtestutils.CheckExportStore(t, info)
	url := "nodelocal://1/listing-test/basepath"
	cloudtestutils.CheckListFiles(
		t, url, username.RootUserName(), nil /*db */, testSettings,
	)
}
