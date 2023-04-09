package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var version = "development"
var commitSha string
var targetOs string
var targetArch string
var buildstamp string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "shows version information for YLS",
	Run: func(cmd *cobra.Command, args []string) {
		type vinfo struct {
			Version    string `json:"version"`
			Commit     string `json:"commit_sha"`
			TargetOS   string `json:"target_os"`
			TargetArch string `json:"target_arch"`
			Timestamp  string `json:"build_timestamp"`
		}

		i := &vinfo{
			Version:    version,
			Commit:     commitSha,
			TargetOS:   targetOs,
			TargetArch: targetArch,
			Timestamp:  buildstamp,
		}
		v, err := json.MarshalIndent(i, "", "    ")
		if err != nil {
			YLSLogger().Fatal("failed to get version info", zap.Error(err))
		}

		fmt.Println(string(v))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
