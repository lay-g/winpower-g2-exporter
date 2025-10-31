package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd 根命令结构
type RootCmd struct {
	cfgFile string
	verbose bool
	cmd     *cobra.Command
}

// NewRootCmd 创建根命令
func NewRootCmd() *RootCmd {
	root := &RootCmd{}

	root.cmd = &cobra.Command{
		Use:   "winpower-g2-exporter",
		Short: "WinPower G2 设备数据采集和 Prometheus 指标导出器",
		Long: `WinPower G2 Exporter 是一个用于采集 WinPower 设备数据、
计算能耗并以 Prometheus 指标形式导出的工具。

支持采集设备状态、电能数据，并提供 HTTP 接口供 Prometheus 抓取。`,
		SilenceUsage:  true,
		SilenceErrors: true,
		// 默认显示帮助信息
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// 添加持久化参数
	root.addPersistentFlags()

	// 添加子命令
	root.cmd.AddCommand(NewServerCmd())
	root.cmd.AddCommand(NewVersionCmd())
	// 注意：Cobra 会自动添加 help 命令，无需手动添加

	return root
}

// addPersistentFlags 添加持久化参数
func (r *RootCmd) addPersistentFlags() {
	r.cmd.PersistentFlags().StringVarP(&r.cfgFile, "config", "c", "",
		"配置文件路径")
	r.cmd.PersistentFlags().BoolVarP(&r.verbose, "verbose", "v", false,
		"详细输出模式")

	// 绑定到 viper
	_ = viper.BindPFlag("config", r.cmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("verbose", r.cmd.PersistentFlags().Lookup("verbose"))
}

// Execute 执行根命令
func (r *RootCmd) Execute() error {
	return r.cmd.Execute()
}

// initConfig 初始化配置
func initConfig(cfgFile string) error {
	if cfgFile != "" {
		// 使用指定的配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 搜索配置文件
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("$HOME/.config/winpower-exporter")
		viper.AddConfigPath("/etc/winpower-exporter")
	}

	// 绑定环境变量
	viper.SetEnvPrefix("WINPOWER_EXPORTER")
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
		// 配置文件不存在不算错误，使用默认配置
	}

	return nil
}
