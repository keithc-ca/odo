package remove

import (
	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/odo/cli/remove/binding"
	"github.com/redhat-developer/odo/pkg/odo/util"
)

// RecommendedCommandName is the recommended remove command name
const RecommendedCommandName = "remove"

// NewCmdDelete implements the odo remove command
func NewCmdRemove(name, fullName string) *cobra.Command {
	var removeCmd = &cobra.Command{
		Use:   name,
		Short: "Remove resources from devfile",
	}

	bindingCmd := binding.NewCmdBinding(binding.BindingRecommendedCommandName, util.GetFullName(fullName, binding.BindingRecommendedCommandName))
	removeCmd.AddCommand(bindingCmd)
	removeCmd.Annotations = map[string]string{"command": "main"}
	removeCmd.SetUsageTemplate(util.CmdUsageTemplate)

	return removeCmd
}