package okteto

import (
	"fmt"
)

// CreateBody top body answer
type CreateBody struct {
	Namespace Namespace `json:"createSpace" yaml:"createSpace"`
}

// DeleteBody top body answer
type DeleteBody struct {
	Namespace Namespace `json:"deleteSpace" yaml:"deleteSpace"`
}

//Namespace represents an Okteto k8s namespace
type Namespace struct {
	ID string `json:"id" yaml:"id"`
}

// CreateNamespace creates a namespace
func CreateNamespace(namespace string) (string, error) {
	q := fmt.Sprintf(`mutation{
		createSpace(name: "%s"){
			id
		},
	}`, namespace)

	var body CreateBody
	if err := query(q, &body); err != nil {
		return "", err
	}

	return body.Namespace.ID, nil
}

// DeleteNamespace deletes a namespace
func DeleteNamespace(namespace string) error {
	q := fmt.Sprintf(`mutation{
		deleteSpace(id: "%s"){
			id
		},
	}`, namespace)

	var body DeleteBody
	if err := query(q, &body); err != nil {
		return err
	}

	return nil
}
