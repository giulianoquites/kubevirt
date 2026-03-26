package clone

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

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

	// Optional flags
	cmd.Flags().StringSlice("label-filter", nil, "Label filters")
	cmd.Flags().StringSlice("annotation-filter", nil, "Annotation filters")

	cmd.Flags().StringSlice("template-label-filter", nil, "Template label filters")
	cmd.Flags().StringSlice("template-annotation-filter", nil, "Template annotation filters")

	cmd.Flags().StringToString("mac", nil, "New MAC addresses (ex: eth0=02:00:00:aa:bb:cc)")
	cmd.Flags().String("serial", "", "New SMBIOS serial")

	cmd.Flags().StringSlice("patch", nil, "JSON patches")

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

	// Read optional flags
	labelFilters, _ := cmd.Flags().GetStringSlice("label-filter")
	annotationFilters, _ := cmd.Flags().GetStringSlice("annotation-filter")

	tmplLabelFilters, _ := cmd.Flags().GetStringSlice("template-label-filter")
	tmplAnnotationFilters, _ := cmd.Flags().GetStringSlice("template-annotation-filter")

	macs, _ := cmd.Flags().GetStringToString("mac")
	serial, _ := cmd.Flags().GetString("serial")

	patches, _ := cmd.Flags().GetStringSlice("patch")

	// Base spec (minimal and safe)
	spec := clonev1.VirtualMachineCloneSpec{
		Source: &corev1.TypedLocalObjectReference{
			APIGroup: pointer.String("kubevirt.io"),
			Kind:     "VirtualMachine",
			Name:     source,
		},
		Target: &corev1.TypedLocalObjectReference{
			APIGroup: pointer.String("kubevirt.io"),
			Kind:     "VirtualMachine",
			Name:     target,
		},
	}

	// Optional fields

	if len(labelFilters) > 0 {
		spec.LabelFilters = labelFilters
	}

	if len(annotationFilters) > 0 {
		spec.AnnotationFilters = annotationFilters
	}

	if len(tmplLabelFilters) > 0 || len(tmplAnnotationFilters) > 0 {
		spec.Template = &clonev1.VirtualMachineCloneTemplateFilters{
			LabelFilters:      tmplLabelFilters,
			AnnotationFilters: tmplAnnotationFilters,
		}
	}

	if len(macs) > 0 {
		spec.NewMacAddresses = macs
	}

	if serial != "" {
		spec.NewSMBiosSerial = pointer.String(serial)
	}

	if len(patches) > 0 {
		spec.Patches = patches
	}

	vmClone := &clonev1.VirtualMachineClone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}

	result, err := virtClient.VirtualMachineClone(namespace).
		Create(context.Background(), vmClone, metav1.CreateOptions{})

	if err != nil {
		return fmt.Errorf("failed to create clone: %v", err)
	}

	fmt.Printf("VirtualMachineClone '%s' created successfully!\n", result.Name)
	return nil
}
