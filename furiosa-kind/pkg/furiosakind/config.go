package furiosakind

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v2"
	kind "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

type Config struct {
	*kind.Cluster
	stdout io.Writer
	stderr io.Writer
}

type ConfigOptions struct {
	defaultName        string
	image              string
	stdout             io.Writer
	stderr             io.Writer
	extraFuncMap       template.FuncMap
	configTemplatePath string
	configTemplate     []byte
	configValuesPath   string
	configValues       []byte
}

type ConfigOption func(*ConfigOptions)

func NewConfig(opts ...ConfigOption) (*Config, error) {
	o := ConfigOptions{}
	for _, opt := range opts {
		opt(&o)
	}
	if o.defaultName == "" {
		o.defaultName = fmt.Sprintf("furiosa-kind-%s", rand.String(5))
	}
	if o.stdout == nil {
		o.stdout = os.Stdout
	}
	if o.stderr == nil {
		o.stderr = os.Stderr
	}
	if o.configTemplate == nil && o.configTemplatePath == "" {
		o.configTemplate = defaultConfigTemplate
	}
	if o.configValues == nil && o.configValuesPath == "" {
		o.configValues = defaultConfigValues
	}
	if o.configTemplate == nil && o.configTemplatePath != "" {
		data, err := os.ReadFile(o.configTemplatePath)
		if err != nil {
			return nil, fmt.Errorf("reading file: %w", err)
		}
		o.configTemplate = data
	}
	if o.configValues == nil && o.configValuesPath != "" {
		data, err := os.ReadFile(o.configValuesPath)
		if err != nil {
			return nil, fmt.Errorf("reading file: %w", err)
		}
		o.configValues = data
	}

	tmpl, err := template.New("configTemplate").Funcs(o.buildFuncMap()).Parse(string(o.configTemplate))
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	var values any
	if err := yaml.Unmarshal(o.configValues, &values); err != nil {
		return nil, fmt.Errorf("unmarshaling YAML: %w", err)
	}
	values = convertToMap(values)

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, values); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	var cluster kind.Cluster
	if err := yaml.Unmarshal(buffer.Bytes(), &cluster); err != nil {
		return nil, fmt.Errorf("unmarshaling YAML: %w", err)
	}

	if cluster.Name == "" {
		cluster.Name = o.defaultName
	}

	if o.image != "" {
		for i := range cluster.Nodes {
			cluster.Nodes[i].Image = o.image
		}
	}

	config := &Config{
		Cluster: &cluster,
		stdout:  o.stdout,
		stderr:  o.stderr,
	}

	return config, nil
}

func (o *ConfigOptions) buildFuncMap() template.FuncMap {
	funcmap := map[string]any{}
	for k, v := range o.extraFuncMap {
		funcmap[k] = v
	}
	for k, v := range sprig.FuncMap() {
		funcmap[k] = v
	}
	return funcmap
}

func convertToMap(data any) any {
	switch v := data.(type) {
	case map[any]any:
		result := make(map[string]any)
		for key, val := range v {
			result[key.(string)] = convertToMap(val)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = convertToMap(item)
		}
		return result
	default:
		return v
	}
}
