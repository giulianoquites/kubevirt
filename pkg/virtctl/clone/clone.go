package clone

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clonev1 "kubevirt.io/api/clone/v1beta1"
	"kubevirt.io/kubevirt/pkg/virtctl/clientconfig"
)

func NewCommand() *cobra.Command {

	var (
		name   string
		source string
		target string

		macMap                    map[string]string
		serial                    string
		labelFilters              []string
		annotationFilters         []string
		templateLabelFilters      []string
		templateAnnotationFilters []string
		patches                   []string
	)

	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone a VirtualMachine",
		Long:  "Clone a VirtualMachine using a VirtualMachineClone resource",

		Example: `
virtctl clone --source vm1 --target vm2
virtctl clone --source vm1 --target vm2 --mac eth0=02:00:00:aa:bb:cc
`,

		RunE: func(cmd *cobra.Command, args []string) error {

			virtClient, namespace, _, err := clientconfig.ClientAndNamespaceFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if name == "" {
				name = fmt.Sprintf("clone-%s", source)
			}

			vmClone := &clonev1.VirtualMachineClone{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: clonev1.VirtualMachineCloneSpec{
					Source: &v1.TypedLocalObjectReference{
						APIGroup: strPtr("kubevirt.io"),
						Kind:     "VirtualMachine",
						Name:     source,
					},
					Target: &v1.TypedLocalObjectReference{
						APIGroup: strPtr("kubevirt.io"),
						Kind:     "VirtualMachine",
						Name:     target,
					},
				},
			}

			// opcionais
			if len(labelFilters) > 0 {
				vmClone.Spec.LabelFilters = labelFilters
			}
			if len(annotationFilters) > 0 {
				vmClone.Spec.AnnotationFilters = annotationFilters
			}
			if len(templateLabelFilters) > 0 || len(templateAnnotationFilters) > 0 {
				vmClone.Spec.Template = clonev1.VirtualMachineCloneTemplateFilters{
					LabelFilters:      templateLabelFilters,
					AnnotationFilters: templateAnnotationFilters,
				}
			}
			if len(macMap) > 0 {
				vmClone.Spec.NewMacAddresses = macMap
			}
			if serial != "" {
				vmClone.Spec.NewSMBiosSerial = &serial
			}
			if len(patches) > 0 {
				vmClone.Spec.Patches = patches
			}

			err = virtClient.RestClient().Post().
				Resource("virtualmachineclones").
				Namespace(namespace).
				Body(vmClone).
				Do(context.Background()).
				Error()

			if err != nil {
				return fmt.Errorf("failed to create clone: %v", err)
			}

			fmt.Printf("Clone '%s' created\n", name)
			return nil
		},
	}

	// 🔥 FLAGS DEFINIDAS AQUI (ESSENCIAL)

	cmd.Flags().StringVar(&name, "name", "", "Clone resource name")
	cmd.Flags().StringVar(&source, "source", "", "Source VM name")
	cmd.Flags().StringVar(&target, "target", "", "Target VM name")

	cmd.Flags().StringToStringVar(&macMap, "mac", nil, "New MAC addresses (eth0=xx:xx:xx)")
	cmd.Flags().StringVar(&serial, "serial", "", "New SMBIOS serial")

	cmd.Flags().StringSliceVar(&labelFilters, "label-filter", nil, "Label filters")
	cmd.Flags().StringSliceVar(&annotationFilters, "annotation-filter", nil, "Annotation filters")

	cmd.Flags().StringSliceVar(&templateLabelFilters, "template-label-filter", nil, "Template label filters")
	cmd.Flags().StringSliceVar(&templateAnnotationFilters, "template-annotation-filter", nil, "Template annotation filters")

	cmd.Flags().StringSliceVar(&patches, "patch", nil, "JSON patches")

	// REQUIRED
	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("target")

	return cmd
}

func strPtr(s string) *string {
	return &s
}
