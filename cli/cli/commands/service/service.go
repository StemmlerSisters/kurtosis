/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/service/logs"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/service/shell"
	"github.com/spf13/cobra"
)

var ServiceCmd = &cobra.Command{
	Use:   command_str_consts.ServiceCmdStr,
	Short: "Manage services",
	RunE:  nil,
}

func init() {
	ServiceCmd.AddCommand(logs.LogsCmd)
	ServiceCmd.AddCommand(shell.ShellCmd)
}
