package secrets

import (
	"fmt"
	"strings"

	"github.com/okteto/okteto/pkg/log"
	"github.com/okteto/okteto/pkg/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	oktetoSecretTemplate = "okteto-%s"
)

// Get returns the value of a secret
func Get(name, namespace string, c *kubernetes.Clientset) (*v1.Secret, error) {
	secret, err := c.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return secret, fmt.Errorf("Error getting kubernetes secret: %s", err)
	}
	return secret, nil
}

//Create creates the syncthing config secret
func Create(dev *model.Dev, c *kubernetes.Clientset) error {
	secretName := GetSecretName(dev)
	log.Debugf("creating configuration secret %s", secretName)

	sct, err := Get(secretName, dev.Namespace, c)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("error getting kubernetes secret: %s", err)
	}

	config, err := getConfigXML(dev)
	if err != nil {
		return fmt.Errorf("error generating syncthing configuration: %s", err)
	}
	data := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName},
		Type:       v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"config.xml": config,
			"cert.pem":   []byte(certPEM),
			"key.pem":    []byte(keyPEM),
		},
	}
	if sct.Name == "" {
		_, err := c.CoreV1().Secrets(dev.Namespace).Create(data)
		if err != nil {
			return fmt.Errorf("error creating kubernetes sync secret: %s", err)
		}

		log.Infof("created okteto secret '%s'.", secretName)
	} else {
		_, err := c.CoreV1().Secrets(dev.Namespace).Update(data)
		if err != nil {
			return fmt.Errorf("error updating kubernetes okteto secret: %s", err)
		}
		log.Infof("okteto secret '%s' was updated.", secretName)
	}
	return nil
}

//Destroy deletes the syncthing config secret
func Destroy(dev *model.Dev, c *kubernetes.Clientset) error {
	secretName := GetSecretName(dev)
	err := c.CoreV1().Secrets(dev.Namespace).Delete(secretName, &metav1.DeleteOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("error deleting kubernetes okteto secret: %s", err)
	}
	return nil
}

//GetSecretName returns the okteto secret name for a given dev environment
func GetSecretName(dev *model.Dev) string {
	return fmt.Sprintf(oktetoSecretTemplate, dev.Name)
}
