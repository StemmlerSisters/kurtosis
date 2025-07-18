package docker_manager

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// writeStaticConfig writes a static Docker config.json file to a temporary directory
func writeStaticConfig(t *testing.T, configContent string) string {
	tmpDir, err := os.MkdirTemp("", "docker-config")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// only write to file if content is not empty
	if configContent != "" {
		configPath := tmpDir + "/config.json"
		err = os.WriteFile(configPath, []byte(configContent), 0600)
		if err != nil {
			t.Fatalf("Failed to write config.json: %v", err)
		}
	}

	// Set the DOCKER_CONFIG environment variable to the temp directory
	os.Setenv(ENV_DOCKER_CONFIG, tmpDir)
	return tmpDir
}

func TestGetAuthWithNoAuthSetReturnsNilAndNoError(t *testing.T) {
	// update docker config env var
	tmpDir := writeStaticConfig(t, "")
	defer os.RemoveAll(tmpDir)
	authConfig, err := GetAuthFromDockerConfig("my-repo/my-image:latest")
	assert.NoError(t, err)
	assert.Nil(t, authConfig, "Auth config should be nil")
}

func TestGetAuthConfigForRepoUserPassword(t *testing.T) {
	expectedUser := "user"
	expectedPassword := "password"

	cfg := fmt.Sprintf(`
	{
		"auths": {
			"https://index.docker.io/v1/": {
				"username": "%s",
				"password": "%s"
			},
			"https://ghcr.io": {
				"password": "pasads"
			},
			"https://example.io": {
				"username": "blabla"
			}
		}
	}`, expectedUser, expectedPassword)

	tmpDir := writeStaticConfig(t, cfg)
	defer os.RemoveAll(tmpDir)

	authConfig, err := GetAuthFromDockerConfig("ubuntu:latest")
	assert.NoError(t, err)
	assert.NotNil(t, authConfig, "Auth config should not be nil")
	assert.Equal(t, expectedUser, authConfig.Username, "Username should match")
	assert.Equal(t, expectedPassword, authConfig.Password, "Password should match")
	assert.Equal(t, "https://index.docker.io/v1/", authConfig.ServerAddress, "Server address should match")

	authConfig, err = GetAuthFromDockerConfig("ghcr.io/my-repo/my-image:latest")
	assert.Nil(t, authConfig, "Auth config should be nil")
	assert.Error(t, err)

	authConfig, err = GetAuthFromDockerConfig("example.io/my-repo/my-image:latest")
	assert.Nil(t, authConfig, "Auth config should be nil")
	assert.Error(t, err)
}

func TestGetAuthConfigForRepoBase64Auth(t *testing.T) {
	expectedUserDockerHub := "dhuser"
	expectedPasswordDockerHub := "dhpassword"
	expectedUserGithub := "ghuser"
	expectedPasswordGithub := "ghpassword"

	encodedAuthDockerHub := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", expectedUserDockerHub, expectedPasswordDockerHub)))
	encodedAuthGithub := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", expectedUserGithub, expectedPasswordGithub)))

	cfg := fmt.Sprintf(`
	{
		"auths": {
			"https://index.docker.io/v1/": {
				"auth": "%s"
			},
			"https://ghcr.io": {
				"auth": "%s"
			}
		}
	}`, encodedAuthDockerHub, encodedAuthGithub)

	tmpDir := writeStaticConfig(t, cfg)
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		repo                  string
		expectedAuth          string
		expectedUser          string
		expectedPassword      string
		expectedServerAddress string
	}{
		{
			repo:                  "alpine:3.17",
			expectedAuth:          encodedAuthDockerHub,
			expectedUser:          expectedUserDockerHub,
			expectedPassword:      expectedPasswordDockerHub,
			expectedServerAddress: "https://index.docker.io/v1/",
		},
		{
			repo:                  "traefik",
			expectedAuth:          encodedAuthDockerHub,
			expectedUser:          expectedUserDockerHub,
			expectedPassword:      expectedPasswordDockerHub,
			expectedServerAddress: "https://index.docker.io/v1/",
		},
		{
			repo:                  "ubuntu:latest",
			expectedAuth:          encodedAuthDockerHub,
			expectedUser:          expectedUserDockerHub,
			expectedPassword:      expectedPasswordDockerHub,
			expectedServerAddress: "https://index.docker.io/v1/",
		},
		{
			repo:                  "docker.io/my-repo/my-image:latest",
			expectedAuth:          encodedAuthDockerHub,
			expectedUser:          expectedUserDockerHub,
			expectedPassword:      expectedPasswordDockerHub,
			expectedServerAddress: "https://index.docker.io/v1/",
		},
		{
			repo:                  "my-repo/my-image:latest",
			expectedAuth:          encodedAuthDockerHub,
			expectedUser:          expectedUserDockerHub,
			expectedPassword:      expectedPasswordDockerHub,
			expectedServerAddress: "https://index.docker.io/v1/",
		},
		{
			repo:                  "https://registry-1.docker.io/my-repo/my-image:latest",
			expectedAuth:          encodedAuthDockerHub,
			expectedUser:          expectedUserDockerHub,
			expectedPassword:      expectedPasswordDockerHub,
			expectedServerAddress: "https://index.docker.io/v1/",
		},
		{
			repo:                  "https://index.docker.io/v1/",
			expectedAuth:          encodedAuthDockerHub,
			expectedUser:          expectedUserDockerHub,
			expectedPassword:      expectedPasswordDockerHub,
			expectedServerAddress: "https://index.docker.io/v1/",
		},
		{
			repo:                  "https://index.docker.io/v1",
			expectedAuth:          encodedAuthDockerHub,
			expectedUser:          expectedUserDockerHub,
			expectedPassword:      expectedPasswordDockerHub,
			expectedServerAddress: "https://index.docker.io/v1/",
		},
		{
			repo:                  "ghcr.io/my-repo/my-image:latest",
			expectedAuth:          encodedAuthGithub,
			expectedUser:          expectedUserGithub,
			expectedPassword:      expectedPasswordGithub,
			expectedServerAddress: "https://ghcr.io/",
		},
		{
			repo:                  "ghcr.io",
			expectedAuth:          encodedAuthGithub,
			expectedUser:          expectedUserGithub,
			expectedPassword:      expectedPasswordGithub,
			expectedServerAddress: "https://ghcr.io/",
		},
		{
			repo:                  "ghcr.io/",
			expectedAuth:          encodedAuthGithub,
			expectedUser:          expectedUserGithub,
			expectedPassword:      expectedPasswordGithub,
			expectedServerAddress: "https://ghcr.io/",
		},
	}

	for _, testCase := range testCases {
		authConfig, err := GetAuthFromDockerConfig(testCase.repo)
		assert.NoError(t, err)
		assert.NotNil(t, authConfig, "Auth config should not be nil")
		assert.Equal(t, testCase.expectedAuth, authConfig.Auth, "Auth for Docker Hub should match")
		assert.Equal(t, testCase.expectedUser, authConfig.Username, "Username should match")
		assert.Equal(t, testCase.expectedPassword, authConfig.Password, "Password should match")
		assert.Equal(t, testCase.expectedServerAddress, authConfig.ServerAddress, "Server address should match")
	}

	authConfig, err := GetAuthFromDockerConfig("something-else.local")
	assert.NoError(t, err)
	assert.Nil(t, authConfig, "Auth config should be nil")

	registries, err := GetAllRegistriesFromDockerConfig()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(registries))
	assert.Contains(t, registries, "https://index.docker.io/v1")
	assert.Contains(t, registries, "https://ghcr.io")
}

func TestGetAuthConfigForRepoOSX(t *testing.T) {
	t.Skip("Skipping test that requires macOS keychain")

	cfg := `{
		"auths": {
			"https://index.docker.io/v1/": {}
		},
		"credsStore": "osxkeychain"
	}`
	tmpDir := writeStaticConfig(t, cfg)
	defer os.RemoveAll(tmpDir)

	authConfig, err := GetAuthFromDockerConfig("my-repo/my-image:latest")
	assert.NoError(t, err)
	assert.NotNil(t, authConfig, "Auth config should not be nil")
}

func TestGetAuthConfigForRepoUnix(t *testing.T) {
	t.Skip("Skipping test that requires unix `pass` password manager")

	cfg := `{
		"auths": {
			"https://index.docker.io/v1/": {}
		},
		"credsStore": "pass"
	}`
	tmpDir := writeStaticConfig(t, cfg)
	defer os.RemoveAll(tmpDir)

	authConfig, err := GetAuthFromDockerConfig("my-repo/my-image:latest")
	assert.NoError(t, err)
	assert.NotNil(t, authConfig, "Auth config should not be nil")
}
