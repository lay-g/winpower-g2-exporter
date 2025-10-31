// Package main 是 WinPower G2 Exporter 的主程序入口点。
package main

import (
	"fmt"
	"os"
)

// 编译时注入的变量，默认值为 dev
var (
	version   = "dev"
	buildTime = ""
	commitID  = ""
)

func main() {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
