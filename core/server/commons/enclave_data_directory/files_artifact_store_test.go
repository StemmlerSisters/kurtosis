/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileStore_StoreFileSavesFile(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	filesArtifactUuid, err := fileStore.StoreFile(reader)
	require.Nil(t, err)
	require.Equal(t, 36, len(filesArtifactUuid)) //UUID is 128 bits but in string it is hex represented chars so 32 chars

	//Test that it saved where it said it would.
	expectedFilename := strings.Join(
		[]string{string(filesArtifactUuid), artifactExtension},
		".",
	)
	expectedFilepath := filepath.Join(fileStore.fileCache.absoluteDirpath, expectedFilename)
	_, dirErr := os.Stat(expectedFilepath)
	require.Nil(t, dirErr)
	file, readErr := ioutil.ReadFile(expectedFilepath)
	require.Nil(t, readErr)
	require.Equal(t, []byte(testContent), file)
}

func TestFileStore_StoreFileToArtifactUUIDSimpleCase(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	targetArtifactUuid, err := NewFilesArtifactUUID()
	require.Equal(t, 36, len(targetArtifactUuid)) //UUID is 128 bits but in string it is hex represented chars so 32 chars
	require.Nil(t, err)
	filesArtifactUuid, err := fileStore.StoreFileToArtifactUUID(reader, targetArtifactUuid)
	require.Nil(t, err)
	require.Equal(t, targetArtifactUuid, filesArtifactUuid)
}

func TestFileStore_StoringToExistingUUIDFails(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	filesArtifactUuid, err := fileStore.StoreFile(reader)
	require.Nil(t, err)
	require.Equal(t, 36, len(filesArtifactUuid)) //UUID is 128 bits but in string it is hex represented chars so 32 chars

	anotherTestContent := "This one should fail"
	anotherReader := strings.NewReader(anotherTestContent)
	_, err = fileStore.StoreFileToArtifactUUID(anotherReader, filesArtifactUuid)
	require.NotNil(t, err)
}

func TestFileStore_GetFilepathByUUIDProperFilepath(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	uuid, err := fileStore.StoreFile(reader)
	require.Nil(t, err)

	enclaveDataFile, err := fileStore.GetFile(uuid)
	require.Nil(t, err)

	_, dirErr := os.Stat(enclaveDataFile.absoluteFilepath)
	require.Nil(t, dirErr)
	file, readErr := ioutil.ReadFile(enclaveDataFile.absoluteFilepath)
	require.Nil(t, readErr)
	require.Equal(t, []byte(testContent), file)
}

func TestFileStore_StoreFilesUniquely(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	otherTestContent := "Long Live Kurtosis, But Different!"

	//Write Both Files
	reader := strings.NewReader(testContent)
	uuid, err := fileStore.StoreFile(reader)
	require.Nil(t, err)

	reader = strings.NewReader(otherTestContent)
	anotherUUID, err := fileStore.StoreFile(reader)
	require.Nil(t, err)
	require.NotEqual(t, uuid, anotherUUID)

	//Get their paths.
	enclaveDataFile, err := fileStore.GetFile(uuid)
	require.Nil(t, err)
	anotherFilepath, err := fileStore.GetFile(anotherUUID)
	require.Nil(t, err)
	require.NotEqual(t, enclaveDataFile, anotherFilepath)

	//Read and evaluate their content is different.
	file, readErr := ioutil.ReadFile(enclaveDataFile.absoluteFilepath)
	require.Nil(t, readErr)
	anotherFile, readErr := ioutil.ReadFile(anotherFilepath.absoluteFilepath)
	require.Nil(t, readErr)
	require.NotEqual(t, file, anotherFile)
}

func TestFileStore_RemoveFileRemovesFileFromDisk(t *testing.T) {
	fileStore := getTestFileStore(t)
	testContent := "Long Live Kurtosis!"
	reader := strings.NewReader(testContent)
	uuid, err := fileStore.StoreFile(reader)
	require.Nil(t, err)

	enclaveDataFile, err := fileStore.GetFile(uuid)
	require.Nil(t, err)

	err = fileStore.RemoveFile(uuid)
	require.Nil(t, err)

	_, err = os.Stat(enclaveDataFile.absoluteFilepath)
	require.NotNil(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestFileStore_RemoveFileFailsForNonExistentUuid(t *testing.T) {
	fileStore := getTestFileStore(t)
	nonExistentUuid, err := NewFilesArtifactUUID()
	require.Nil(t, err)

	err = fileStore.RemoveFile(nonExistentUuid)
	require.NotNil(t, err)
}

func getTestFileStore(t *testing.T) *FilesArtifactStore {
	absDirpath, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	fileStore := newFilesArtifactStore(absDirpath, "")
	require.Nil(t, err)
	return fileStore
}
