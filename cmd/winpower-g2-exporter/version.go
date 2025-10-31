package main

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// VersionInfo 版本信息结构
type VersionInfo struct {
	Version   string `json:"version"`    // 版本号
	GoVersion string `json:"go_version"` // Go 运行时版本
	BuildTime string `json:"build_time"` // 编译时间
	CommitID  string `json:"commit_id"`  // Commit ID
	Platform  string `json:"platform"`   // 运行平台
	Compiler  string `json:"compiler"`   // 编译器信息
}

// NewVersionCmd 创建 version 子命令
func NewVersionCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Long: `显示应用程序的版本信息，包括：
- 版本号
- Go 运行时信息
- 编译时间
- Git Commit ID
- 平台信息`,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := getVersionInfo()

			switch format {
			case "json":
				return outputJSON(info)
			default:
				return outputText(info)
			}
		},
	}

	// 添加输出格式参数
	cmd.Flags().StringVarP(&format, "format", "f", "text",
		"输出格式 (text|json)")

	return cmd
}

// getVersionInfo 获取版本信息
func getVersionInfo() *VersionInfo {
	return &VersionInfo{
		Version:   version,
		GoVersion: runtime.Version(),
		BuildTime: buildTime,
		CommitID:  commitID,
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
		Compiler:  runtime.Compiler,
	}
}

// outputJSON 以 JSON 格式输出
func outputJSON(info *VersionInfo) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化版本信息失败: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// outputText 以文本格式输出
func outputText(info *VersionInfo) error {
	fmt.Printf("WinPower G2 Exporter\n")
	fmt.Printf("  Version:    %s\n", info.Version)
	fmt.Printf("  Go Version: %s\n", info.GoVersion)
	fmt.Printf("  Build Time: %s\n", info.BuildTime)
	fmt.Printf("  Commit ID:  %s\n", info.CommitID)
	fmt.Printf("  Platform:   %s\n", info.Platform)
	fmt.Printf("  Compiler:   %s\n", info.Compiler)
	return nil
}
