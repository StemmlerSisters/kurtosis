/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_volume

import (
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

func TestGetFile(t *testing.T) {
	enclaveDirpath, err := ioutil.TempDir("", "")
	assert.Nil(t, err)

	enclaveDir := NewEnclaveDataVolume(enclaveDirpath)

	serviceGUID := service_network_types.ServiceGUID("someService")

	svcDir, err := enclaveDir.GetServiceDirectory(serviceGUID)
	assert.Nil(t, err)

	svcAbsDirpath := svcDir.absoluteDirpath
	svcRelDirpath := svcDir.dirpathRelativeToVolRoot

	filename := "someFile"

	file, err := svcDir.NewGeneratedFile(filename)
	assert.Nil(t, err)

	// Check file was actually created
	files, err := ioutil.ReadDir(svcAbsDirpath)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(files))
	fileInfo := files[0]
	assert.True(t, strings.Contains(fileInfo.Name(), filename))

	// Check EnclaveDataVolFile data structure is correct
	assert.Equal(
		t,
		svcAbsDirpath,
		path.Dir(file.GetAbsoluteFilepath()),
	)
	assert.Equal(
		t,
		svcRelDirpath,
		path.Dir(file.GetFilepathRelativeToVolRoot()),
	)
}
