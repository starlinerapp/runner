package handler

import (
	"fmt"

	"github.com/spf13/cobra"
	"starliner.app/runner/internal/application"
	"starliner.app/runner/internal/infrastructure/firecracker/config"
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
		vh.newListVMCmd(),
		vh.newDeleteVMCmd(),
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

func (vh *VMHandler) newListVMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List virtual machines",
		Long:  "List virtual machines",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			vms, err := vh.vmApplication.ListVMs()
			if err != nil {
				return err
			}
			for _, vm := range vms {
				fmt.Printf("VM %s\n", vm.ID)
				fmt.Printf("  tap:         %s\n", vm.Tap)
				fmt.Printf("  mac:         %s\n", vm.MAC)
				fmt.Printf("  subnet:      172.16.%d.0/24\n", vm.SubnetOctet)
				fmt.Printf("  guest cid:   %d\n", vm.GuestCID)
				fmt.Printf("  workspace:   %s\n", vm.Dir)
				fmt.Printf("  log:         %s\n", config.LogPath(vm.Dir))
				fmt.Printf("  firecracker: pid %d\n", vm.FirecrackerPID)
				fmt.Printf("  created:     %s\n", vm.CreatedAt.Format("2006-01-02 15:04:05 UTC"))
				fmt.Println()
			}
			return nil
		},
	}
	return cmd
}

func (vh *VMHandler) newDeleteVMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a virtual machine",
		Long:  "Delete a virtual machine",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id := args[0]
			if err := vh.vmApplication.DeleteVM(id); err != nil {
				return err
			}
			fmt.Printf("VM %s deleted\n", id)
			return nil
		},
	}
	return cmd
}
