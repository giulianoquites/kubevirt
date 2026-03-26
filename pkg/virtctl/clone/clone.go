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

// ----------------------------
// Command
// ----------------------------

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone a VirtualMachine",
		Long:  "Clone a VirtualMachine using a VirtualMachineClone resource",

		Example: `
# Basic clone
virtctl clone --source vm1 --target vm2

# Clone with MAC address
virtctl clone --source vm1 --target vm2 --mac eth0=02:00:00:aa:bb:cc
`,

		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd)
		},
	}

	// Required
	cmd.Flags().String("name", "", "Clone resource name")
	cmd.Flags().String("source", "", "Source VM name")
	cmd.Flags().String("target", "", "Target VM name")

	// Optional
	cmd.Flags().StringToString("mac", nil, "New MAC addresses (e.g. eth0=02:00:00:aa:bb:cc)")
	cmd.Flags().String("serial", "", "New SMBIOS serial")

	cmd.Flags().StringSlice("label-filter", nil, "Label filters")
	cmd.Flags().StringSlice("annotation-filter", nil, "Annotation filters")

	cmd.Flags().StringSlice("template-label-filter", nil, "Template label filters")
	cmd.Flags().StringSlice("template-annotation-filter", nil, "Template annotation filters")

	cmd.Flags().StringSlice("patch", nil, "JSON patches")

	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("target")

	return cmd
}

// ----------------------------
// Run
// ----------------------------

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

	// Optional flags
	macMap, _ := cmd.Flags().GetStringToString("mac")
	serial, _ := cmd.Flags().GetString("serial")

	labelFilters, _ := cmd.Flags().GetStringSlice("label-filter")
	annotationFilters, _ := cmd.Flags().GetStringSlice("annotation-filter")

	templateLabelFilters, _ := cmd.Flags().GetStringSlice("template-label-filter")
	templateAnnotationFilters, _ := cmd.Flags().GetStringSlice("template-annotation-filter")

	patches, _ := cmd.Flags().GetStringSlice("patch")

	// ----------------------------
	// Build Spec
	// ----------------------------

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

	// Optional fields

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

	// ----------------------------
	// Create Clone
	// ----------------------------

	result := &clonev1.VirtualMachineClone{}
	err = virtClient.RestClient().Post().
		Resource("virtualmachineclones").
		Namespace(namespace).
		Body(vmClone).
		Do(context.Background()).
		Into(result)

	if err != nil {
		return fmt.Errorf("failed to create clone: %v", err)
	}

	fmt.Printf("VirtualMachineClone '%s' created successfully\n", result.Name)
	return nil
}

// ----------------------------
// Helpers
// ----------------------------

func strPtr(s string) *string {
	return &s
}
