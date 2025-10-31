package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCmd(t *testing.T) {
	root := NewRootCmd()

	assert.NotNil(t, root)
	assert.NotNil(t, root.cmd)
	assert.Equal(t, "winpower-g2-exporter", root.cmd.Use)
}

func TestRootCmdExecute(t *testing.T) {
	root := NewRootCmd()

	// 测试不带参数执行（应该显示帮助）
	root.cmd.SetArgs([]string{})
	err := root.Execute()

	// 帮助命令不应该返回错误
	assert.NoError(t, err)
}

func TestRootCmdHasSubcommands(t *testing.T) {
	root := NewRootCmd()

	commands := root.cmd.Commands()

	// 应该有 server 和 version 子命令
	commandNames := make([]string, 0, len(commands))
	for _, cmd := range commands {
		commandNames = append(commandNames, cmd.Use)
	}

	// 验证必需的子命令存在
	assert.Contains(t, commandNames, "server")
	assert.Contains(t, commandNames, "version")
	// Cobra 会自动添加 help 和 completion 命令
	assert.GreaterOrEqual(t, len(commandNames), 2, "应该至少有 server 和 version 两个子命令")
}
