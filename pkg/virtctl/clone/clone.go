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

type cloneOptions struct {
	namespace string
	source    string
	target    string

	labelFilters      []string
	annotationFilters []string

	templateLabelFilters      []string
	templateAnnotationFilters []string

	newMacAddresses map[string]string
	newSMBiosSerial string

	patches []string
}

func NewCommand() *cobra.Command {
	opts := &cloneOptions{
		newMacAddresses: map[string]string{},
	}

	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone a VirtualMachine using a VirtualMachineClone resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.source, "source", "", "Source VM name")
	cmd.Flags().StringVar(&opts.target, "target", "", "Target VM name")

	cmd.Flags().StringSliceVar(&opts.labelFilters, "labels", nil, "Label filters")
	cmd.Flags().StringSliceVar(&opts.annotationFilters, "annotations", nil, "Annotation filters")

	cmd.Flags().StringSliceVar(&opts.templateLabelFilters, "template-labels", nil, "Template label filters")
	cmd.Flags().StringSliceVar(&opts.templateAnnotationFilters, "template-annotations", nil, "Template annotation filters")

	cmd.Flags().StringToStringVar(&opts.newMacAddresses, "mac", nil, "MAC addresses (eth0=xx:xx:xx)")
	cmd.Flags().StringVar(&opts.newSMBiosSerial, "serial", "", "SMBIOS serial")

	cmd.Flags().StringSliceVar(&opts.patches, "patch", nil, "JSON patches")

	cmd.Flags().StringVarP(&opts.namespace, "namespace", "n", "", "Namespace")

	_ = cmd.MarkFlagRequired("source")
	_ = cmd.MarkFlagRequired("target")

	return cmd
}

func run(ctx context.Context, opts *cloneOptions) error {
	virtClient, namespace, _, err := clientconfig.ClientAndNamespaceFromContext(ctx)
	if err != nil {
		return err
	}

	if opts.namespace != "" {
		namespace = opts.namespace
	}

	vmClone := &clonev1.VirtualMachineClone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.target + "-clone",
			Namespace: namespace,
		},
		Spec: clonev1.VirtualMachineCloneSpec{
			Source: &corev1.TypedLocalObjectReference{
				APIGroup: strPtr("kubevirt.io"),
				Kind:     "VirtualMachine",
				Name:     opts.source,
			},
			Target: &corev1.TypedLocalObjectReference{
				APIGroup: strPtr("kubevirt.io"),
				Kind:     "VirtualMachine",
				Name:     opts.target,
			},
		},
	}

	// opcionais
	if len(opts.labelFilters) > 0 {
		vmClone.Spec.LabelFilters = opts.labelFilters
	}

	if len(opts.annotationFilters) > 0 {
		vmClone.Spec.AnnotationFilters = opts.annotationFilters
	}

	if len(opts.templateLabelFilters) > 0 || len(opts.templateAnnotationFilters) > 0 {
		vmClone.Spec.Template = clonev1.VirtualMachineCloneTemplateFilters{
			LabelFilters:      opts.templateLabelFilters,
			AnnotationFilters: opts.templateAnnotationFilters,
		}
	}

	if len(opts.newMacAddresses) > 0 {
		vmClone.Spec.NewMacAddresses = opts.newMacAddresses
	}

	if opts.newSMBiosSerial != "" {
		vmClone.Spec.NewSMBiosSerial = strPtr(opts.newSMBiosSerial)
	}

	if len(opts.patches) > 0 {
		vmClone.Spec.Patches = opts.patches
	}

	_, err = virtClient.VirtualMachineClone(namespace).Create(ctx, vmClone, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create clone: %v", err)
	}

	fmt.Printf("Clone %s created successfully\n", vmClone.Name)
	return nil
}

func strPtr(s string) *string {
	return &s
}
