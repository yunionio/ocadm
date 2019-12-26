package addons

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/pkg/errors"

	"yunion.io/x/ocadm/pkg/util/kubectl"
)

func CompileTemplateFromMap(tmplt string, configMap interface{}) (string, error) {
	out := new(bytes.Buffer)
	t := template.Must(template.New("compiled_template").Parse(tmplt))
	if err := t.Execute(out, configMap); err != nil {
		return "", err
	}
	return out.String(), nil
}

type Configer interface {
	Name() string
	GenerateYAML() (string, error)
}

func KubectlApplyAddon(c Configer, client *kubectl.Client, onlyShow bool) error {
	manifest, err := c.GenerateYAML()
	if err != nil {
		return errors.Wrapf(err, "get addon %s manifest", c.Name())
	}
	if onlyShow {
		fmt.Printf("%s", manifest)
		return nil
	}
	if err := client.Apply(manifest); err != nil {
		return errors.Wrapf(err, "apply addon %s", c.Name())
	}
	fmt.Printf("[oc-addons] Applied addon: %s\n", c.Name())
	return nil
}
