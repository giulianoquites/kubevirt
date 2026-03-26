package clone

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clonev1 "kubevirt.io/api/clone/v1beta1"
	"kubevirt.io/kubevirt/pkg/virtctl/clientconfig"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone a VirtualMachine.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd)
		},
	}

	cmd.Flags().String("name", "", "Name of the clone resource")
	cmd.Flags().String("source", "", "Source VM name")
	cmd.Flags().String("target", "", "Target VM name")

	_ = cmd.MarkFlagRequired("source")
	_ = cmd.MarkFlagRequired("target")

	return cmd
}

func run(cmd *cobra.Command) error {
	virtClient, namespace, _, err := clientconfig.ClientAndNamespaceFromContext(cmd.Context())
	if err != nil {
		return err
	}

	source, _ := cmd.Flags().GetString("source")
	target, _ := cmd.Flags().GetString("target")
	name, _ := cmd.Flags().GetString("name")

	if name == "" {
		name = fmt.Sprintf("clone-%s", source)
	}

	vmClone := &clonev1.VirtualMachineClone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: clonev1.VirtualMachineCloneSpec{
			Source: &corev1.TypedLocalObjectReference{
				Kind: "VirtualMachine",
				Name: source,
			},
			Target: &corev1.TypedLocalObjectReference{
				Kind: "VirtualMachine",
				Name: target,
			},
		},
	}

	result := &clonev1.VirtualMachineClone{}

	err = virtClient.RestClient().
    	Post().
    	AbsPath(fmt.Sprintf("/apis/clone.kubevirt.io/v1beta1/namespaces/%s/virtualmachineclones", namespace)).
    	Body(vmClone).
   		Do(context.Background()).
    	Into(result)

	if err != nil {
		return fmt.Errorf("failed to create clone: %v", err)
	}

	fmt.Printf("VirtualMachineClone '%s' created successfully!\n", result.Name)
	return nil
}
