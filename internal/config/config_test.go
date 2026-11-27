package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitConfigWithEnvCredentials(t *testing.T) {
	if path := os.Getenv("IMAGEHOST_TEST_CONFIG_PATH"); path != "" {
		InitConfig(path)
		return
	}
	for _, tc := range []struct {
		name     string
		content  string
		wantOut  string
		wantFail bool
	}{
		{"env only", "", "", false},
		{"malformed file", "api_token: [", "读取配置文件", true},
		{"public example token", "api_token: PRIVATE_123_KEY", "默认示例 api_token", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if tc.content != "" {
				if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			// InitConfig exits the process on startup errors.
			cmd := exec.Command(os.Args[0], "-test.run=^TestInitConfigWithEnvCredentials$")
			cmd.Env = append(os.Environ(), "IMAGEHOST_TEST_CONFIG_PATH="+path,
				"CLIENT_ID=test", "CLIENT_SECRET=test", "API_TOKEN=", "PORT=8080")
			out, err := cmd.CombinedOutput()
			if (err != nil) != tc.wantFail || (tc.wantOut != "" && !strings.Contains(string(out), tc.wantOut)) {
				t.Fatalf("unexpected startup result: %v\n%s", err, out)
			}
		})
	}
}
