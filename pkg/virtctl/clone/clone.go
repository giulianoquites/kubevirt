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
		Short: "Clone a VirtualMachine",
		Long: `Clone a VirtualMachine using KubeVirt VirtualMachineClone resource.

This command creates a clone of an existing VirtualMachine into a new one.
It supports optional customization such as labels, annotations, MAC addresses,
SMBIOS serial, and JSON patches.`,
		Example: `
# Basic clone
virtctl clone --source vm1 --target vm2

# Clone with MAC address
virtctl clone --source vm1 --target vm2 \
  --mac eth0=02:00:00:aa:bb:cc

# Clone with filters
virtctl clone --source vm1 --target vm2 \
  --label-filter "*" \
  --annotation-filter "!network/*"

# Clone with everything
virtctl clone \
  --source vm1 \
  --target vm2 \
  --mac eth0=02:00:00:aa:bb:cc \
  --serial my-serial \
  --patch '{"op":"add","path":"/metadata/labels/test","value":"ok"}'
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd)
		},
	}

	// Required flags
	cmd.Flags().String("name", "", "Name of the VirtualMachineClone resource (default: clone-<source>)")
	cmd.Flags().String("source", "", "Source VirtualMachine name")
	cmd.Flags().String("target", "", "Target VirtualMachine name")

	// Optional flags
	cmd.Flags().StringSlice("label-filter", nil, "Label filters to include/exclude (e.g. '*', '!key/*')")
	cmd.Flags().StringSlice("annotation-filter", nil, "Annotation filters to include/exclude")

	cmd.Flags().StringSlice("template-label-filter", nil, "Template label filters")
	cmd.Flags().StringSlice("template-annotation-filter", nil, "Template annotation filters")

	cmd.Flags().StringToString("mac", nil, "New MAC addresses (e.g. eth0=02:00:00:aa:bb:cc)")
	cmd.Flags().String("serial", "", "New SMBIOS serial number")

	cmd.Flags().StringSlice("patch", nil, "JSON patches to customize the cloned VM")

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

	// Optional flags
	labelFilters, _ := cmd.Flags().GetStringSlice("label-filter")
	annotationFilters, _ := cmd.Flags().GetStringSlice("annotation-filter")

	tmplLabelFilters, _ := cmd.Flags().GetStringSlice("template-label-filter")
	tmplAnnotationFilters, _ := cmd.Flags().GetStringSlice("template-annotation-filter")

	macs, _ := cmd.Flags().GetStringToString("mac")
	serial, _ := cmd.Flags().GetString("serial")

	patches, _ := cmd.Flags().GetStringSlice("patch")

	// Base spec
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
		spec.Template = clonev1.VirtualMachineCloneTemplateFilters{
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
