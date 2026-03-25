package clone

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clonev1 "kubevirt.io/api/clone/v1alpha1"

	"kubevirt.io/kubevirt/pkg/virtctl/templates"
)

func NewCloneCommand(clientGetter templates.KubecliGetter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clona uma VirtualMachine.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(clientGetter, cmd)
		},
	}

	// Flags do comando
	cmd.Flags().String("name", "", "Nome do recurso de clone")
	cmd.Flags().String("source", "", "Nome da VM de origem")
	cmd.Flags().String("target", "", "Nome da nova VM (alvo)")
	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("target")

	return cmd
}

func run(clientGetter templates.KubecliGetter, cmd *cobra.Command) error {
	virtClient, err := clientGetter.Get()
	if err != nil {
		return err
	}

	namespace, _, _ := clientGetter.ToRawKubeConfigLoader().Namespace()
	source, _ := cmd.Flags().GetString("source")
	target, _ := cmd.Flags().GetString("target")
	name, _ := cmd.Flags().GetString("name")

	if name == "" {
		name = fmt.Sprintf("clone-%s", source)
	}

	// Definindo o objeto VirtualMachineClone
	vmClone := &clonev1.VirtualMachineClone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: clonev1.VirtualMachineCloneSpec{
			Source: &v1.TypedLocalObjectReference{
				APIGroup: &clonev1.SchemeGroupVersion.Group,
				Kind:     "VirtualMachine",
				Name:     source,
			},
			Target: &v1.TypedLocalObjectReference{
				APIGroup: &clonev1.SchemeGroupVersion.Group,
				Kind:     "VirtualMachine",
				Name:     target,
			},
		},
	}

	// Enviando para a API do Kubernetes
	result := &clonev1.VirtualMachineClone{}
	err = virtClient.RestClient().Post().
		Resource("virtualmachineclones").
		Namespace(namespace).
		Body(vmClone).
		Do(context.Background()).
		Into(result)

	if err != nil {
		return fmt.Errorf("falha ao criar clone: %v", err)
	}

	fmt.Printf("Recurso VirtualMachineClone '%s' criado com sucesso!\n", result.Name)
	return nil
}
