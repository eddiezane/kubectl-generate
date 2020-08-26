package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	openapi_v2 "github.com/googleapis/gnostic/OpenAPIv2"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// GenerateOptions encapsulates configuration info
type GenerateOptions struct {
	ConfigFlags *genericclioptions.ConfigFlags

	ResourceName string

	UpstreamSchema *openapi_v2.Document

	CustomSchema     *openapi_v2.Document
	CustomSchemaPath string

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

	cmd.Flags().StringVar(&o.CustomSchemaPath, "schema", "", "Local file path or URL to load as example schema")
	o.ConfigFlags.AddFlags(cmd.Flags())

	return cmd
}

// Complete prepares required configurations
func (o *GenerateOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("Resource to generate required")
	}
	resourceName := strings.ToLower(args[0])
	o.ResourceName = resourceName

	upstream, err := getUpstreamSchema(o)
	if err != nil {
		return err
	}
	o.UpstreamSchema = upstream

	custom, err := getCustomSchema(o.CustomSchemaPath)
	if err != nil {
		return err
	}
	o.CustomSchema = custom

	return nil
}

// Validate will ensure that configurations provided are acceptable
func (o *GenerateOptions) Validate() error {
	if o.ResourceName != "deployment" {
		return errors.New("Only deployment is currently supported")
	}
	return nil
}

// Run executes the command
func (o *GenerateOptions) Run() error {
	schema := mergeSchema(o.CustomSchema, o.UpstreamSchema)

	var item *openapi_v2.NamedSchema
	for _, i := range schema.GetDefinitions().GetAdditionalProperties() {
		if i.GetName() == "io.k8s.api.apps.v1.Deployment" {
			item = i
			break
		}
	}
	example := item.GetValue().GetExample().ToRawInfo().Value
	fmt.Fprintln(o.Out, example)
	return nil
}

func getUpstreamSchema(o *GenerateOptions) (*openapi_v2.Document, error) {
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	discoveryClient, err := factory.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	schema, err := discoveryClient.OpenAPISchema()
	if err != nil {
		return nil, err
	}
	return schema, nil
}

// getCustomSchema will parse the schema whether it exists locally or remote
func getCustomSchema(path string) (*openapi_v2.Document, error) {
	var schema []byte

	_, err := os.Stat(path)

	if !os.IsNotExist(err) {
		f, err := os.Open(path)
		defer f.Close()

		if err != nil {
			return nil, err
		}

		schema, err = ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
	} else {
		response, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		schema, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
	}

	return openapi_v2.ParseDocument([]byte(schema))
}

func mergeSchema(local, upstream *openapi_v2.Document) *openapi_v2.Document {
	var localExamples = map[string]*openapi_v2.Any{}

	for _, i := range local.GetDefinitions().GetAdditionalProperties() {
		if strings.HasPrefix(i.GetName(), "io.k8s.config.examples/") {
			localExamples[strings.Replace(i.GetName(), "config.examples/", "", 1)] = i.GetValue().GetExample()
		}
	}

	for _, i := range upstream.GetDefinitions().GetAdditionalProperties() {
		if example, found := localExamples[i.GetName()]; found {
			i.GetValue().Example = example
		}
	}

	return upstream
}
