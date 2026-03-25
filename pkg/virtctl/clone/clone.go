package clone

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clonev1 "kubevirt.io/api/clone/v1beta1"

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
	cmd.Flags().String("source-namespace", "", "Namespace da VM de origem (opcional)")
	cmd.Flags().String("target-namespace", "", "Namespace da VM alvo (opcional)")

	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("target")

	return cmd
}

func run(clientGetter templates.KubecliGetter, cmd *cobra.Command) error {
	virtClient, err := clientGetter.Get()
	if err != nil {
		return err
	}

	currentNamespace, _, _ := clientGetter.ToRawKubeConfigLoader().Namespace()

	source, _ := cmd.Flags().GetString("source")
	target, _ := cmd.Flags().GetString("target")
	name, _ := cmd.Flags().GetString("name")
	sourceNamespace, _ := cmd.Flags().GetString("source-namespace")
	targetNamespace, _ := cmd.Flags().GetString("target-namespace")

	if name == "" {
		name = fmt.Sprintf("clone-%s", source)
	}

	if sourceNamespace == "" {
		sourceNamespace = currentNamespace
	}

	if targetNamespace == "" {
		targetNamespace = currentNamespace
	}

	group := "kubevirt.io"

	// Definindo o objeto VirtualMachineClone
	vmClone := &clonev1.VirtualMachineClone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: targetNamespace,
		},
		Spec: clonev1.VirtualMachineCloneSpec{
			Source: &clonev1.TypedObjectReference{
				APIGroup:  &group,
				Kind:      "VirtualMachine",
				Name:      source,
				Namespace: sourceNamespace,
			},
			Target: &clonev1.TypedObjectReference{
				APIGroup:  &group,
				Kind:      "VirtualMachine",
				Name:      target,
				Namespace: targetNamespace,
			},
		},
	}

	// Enviando para a API do Kubernetes
	result := &clonev1.VirtualMachineClone{}
	err = virtClient.RestClient().Post().
		Resource("virtualmachineclones").
		Namespace(targetNamespace).
		Body(vmClone).
		Do(context.Background()).
		Into(result)

	if err != nil {
		return fmt.Errorf("falha ao criar clone: %v", err)
	}

	fmt.Printf("Recurso VirtualMachineClone '%s' criado com sucesso!\n", result.Name)
	return nil
}
