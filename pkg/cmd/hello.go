package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// GenerateOptions encapsulates configuration info
type GenerateOptions struct {
	ConfigFlags *genericclioptions.ConfigFlags

	Addressee string

	genericclioptions.IOStreams
}

// NewGenerateOptions will return an instance of GenerateOptions
func NewGenerateOptions(streams genericclioptions.IOStreams) *GenerateOptions {
	return &GenerateOptions{
		ConfigFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

// NewCmdGenerate creates and returns a command to generate resource manifests.
func NewCmdGenerate(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewGenerateOptions(streams)

	cmd := &cobra.Command{
		Use:   "kubectl generate",
		Short: "prints Kubernetes resource manifests",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&o.Addressee, "addressee", "World", "person/entity to greet")
	o.ConfigFlags.AddFlags(cmd.Flags())

	return cmd
}

// Run executes the command
func (o *GenerateOptions) Run() error {
	fmt.Printf("Hello, %s!\n", o.Addressee)
	return nil
}

// Validate will ensure that configurations provided are acceptable
func (o *GenerateOptions) Validate() error {
	return nil
}

// Complete prepares required configurations
func (o *GenerateOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}
