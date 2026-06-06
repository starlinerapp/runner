package handler

import (
	"github.com/spf13/cobra"
	"starliner.app/runner/internal/application"
)

type VMHandler struct {
	vmApplication *application.VMApplication
}

func NewVMHandler(
	vmApplication *application.VMApplication,
) *VMHandler {
	return &VMHandler{
		vmApplication: vmApplication,
	}
}

func (vh *VMHandler) NewVMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vm",
		Short: "Manage your virtual machines",
		Long:  "Manage your virtual machines",
	}

	cmd.AddCommand(
		vh.newCreateVMCmd(),
	)

	return cmd
}

func (vh *VMHandler) newCreateVMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new virtual machine",
		Long:  "Create a new virtual machine",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return vh.vmApplication.CreateVM()
		},
	}
	return cmd
}
