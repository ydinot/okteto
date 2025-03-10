package deployments

import (
	"reflect"
	"testing"

	"github.com/okteto/okteto/pkg/model"
	yaml "gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

func Test_translate(t *testing.T) {
	var runAsUser int64 = 100
	var runAsGroup int64 = 101
	var fsGroup int64 = 102
	manifest := []byte(`name: web
container: dev
image: web:latest
command: ["./run_web.sh"]
workdir: /app
securityContext:
  runAsUser: 100
  runAsGroup: 101
  fsGroup: 102
services:
  - name: worker
    container: dev
    image: worker:latest
    command: ["./run_worker.sh"]
    mountpath: /src
    subpath: /worker`)

	dev, err := model.Read(manifest)
	if err != nil {
		t.Fatal(err)
	}
	d1 := dev.GevSandbox()
	rule1 := dev.ToTranslationRule(dev, d1, "node")
	tr1 := &model.Translation{
		Interactive: true,
		Name:        dev.Name,
		Deployment:  d1,
		Rules:       []*model.TranslationRule{rule1},
	}
	err = translate(tr1, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	d1OK := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					SecurityContext: &apiv1.PodSecurityContext{
						RunAsUser:  &runAsUser,
						RunAsGroup: &runAsGroup,
						FSGroup:    &fsGroup,
					},
					TerminationGracePeriodSeconds: &devTerminationGracePeriodSeconds,
					Volumes: []apiv1.Volume{
						apiv1.Volume{
							Name: "pvc-0-okteto-web-0",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: "pvc-0-okteto-web-0",
									ReadOnly:  false,
								},
							},
						},
						apiv1.Volume{
							Name: oktetoBinName,
							VolumeSource: apiv1.VolumeSource{
								EmptyDir: &apiv1.EmptyDirVolumeSource{},
							},
						},
					},
					NodeName: "node",
					InitContainers: []apiv1.Container{
						apiv1.Container{
							Name:            oktetoBinName,
							Image:           "okteto/bin",
							ImagePullPolicy: apiv1.PullAlways,
							Command:         []string{"sh", "-c", "cp /usr/local/bin/* /okteto/bin"},
							VolumeMounts: []apiv1.VolumeMount{
								apiv1.VolumeMount{
									Name:      oktetoBinName,
									MountPath: "/okteto/bin",
								},
							},
						},
					},
					Containers: []apiv1.Container{
						apiv1.Container{
							Name:            "dev",
							Image:           "web:latest",
							ImagePullPolicy: apiv1.PullAlways,
							Command:         []string{"tail"},
							Args:            []string{"-f", "/dev/null"},
							WorkingDir:      "/app",
							VolumeMounts: []apiv1.VolumeMount{
								apiv1.VolumeMount{
									Name:      "pvc-0-okteto-web-0",
									ReadOnly:  false,
									MountPath: "/app",
								},
								apiv1.VolumeMount{
									Name:      oktetoBinName,
									ReadOnly:  false,
									MountPath: "/var/okteto/bin",
								},
							},
							LivenessProbe:  nil,
							ReadinessProbe: nil,
						},
					},
				},
			},
		},
	}
	marshalled1, _ := yaml.Marshal(d1.Spec.Template.Spec)
	marshalled1OK, _ := yaml.Marshal(d1OK.Spec.Template.Spec)
	if string(marshalled1) != string(marshalled1OK) {
		t.Fatalf("Wrong d1 generation.\nActual %s, \nExpected %s", string(marshalled1), string(marshalled1OK))
	}

	dev2 := dev.Services[0]
	d2 := dev2.GevSandbox()
	rule2 := dev2.ToTranslationRule(dev, d2, "node")
	tr2 := &model.Translation{
		Interactive: false,
		Name:        dev.Name,
		Deployment:  d2,
		Rules:       []*model.TranslationRule{rule2},
	}
	err = translate(tr2, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	d2OK := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					TerminationGracePeriodSeconds: &devTerminationGracePeriodSeconds,
					Volumes: []apiv1.Volume{
						apiv1.Volume{
							Name: "pvc-0-okteto-web-0",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: "pvc-0-okteto-web-0",
									ReadOnly:  false,
								},
							},
						},
						apiv1.Volume{
							Name: oktetoBinName,
							VolumeSource: apiv1.VolumeSource{
								EmptyDir: &apiv1.EmptyDirVolumeSource{},
							},
						},
					},
					NodeName: "node",
					InitContainers: []apiv1.Container{
						apiv1.Container{
							Name:            oktetoBinName,
							Image:           "okteto/bin",
							ImagePullPolicy: apiv1.PullAlways,
							Command:         []string{"sh", "-c", "cp /usr/local/bin/* /okteto/bin"},
							VolumeMounts: []apiv1.VolumeMount{
								apiv1.VolumeMount{
									Name:      oktetoBinName,
									MountPath: "/okteto/bin",
								},
							},
						},
					},
					Containers: []apiv1.Container{
						apiv1.Container{
							Name:            "dev",
							Image:           "worker:latest",
							ImagePullPolicy: apiv1.PullAlways,
							Command:         []string{"./run_worker.sh"},
							Args:            []string{},
							VolumeMounts: []apiv1.VolumeMount{
								apiv1.VolumeMount{
									Name:      "pvc-0-okteto-web-0",
									ReadOnly:  false,
									MountPath: "/src",
									SubPath:   "/worker",
								},
								apiv1.VolumeMount{
									Name:      oktetoBinName,
									ReadOnly:  false,
									MountPath: "/var/okteto/bin",
								},
							},
							LivenessProbe:  nil,
							ReadinessProbe: nil,
						},
					},
				},
			},
		},
	}
	marshalled2, _ := yaml.Marshal(d2.Spec.Template.Spec)
	marshalled2OK, _ := yaml.Marshal(d2OK.Spec.Template.Spec)
	if string(marshalled2) != string(marshalled2OK) {
		t.Fatalf("Wrong d2 generation.\nActual %s, \nExpected %s", string(marshalled2), string(marshalled2OK))
	}
}

func Test_translateResources(t *testing.T) {
	type args struct {
		c *apiv1.Container
		r model.ResourceRequirements
	}
	tests := []struct {
		name             string
		args             args
		expectedRequests map[apiv1.ResourceName]resource.Quantity
		expectedLimits   map[apiv1.ResourceName]resource.Quantity
	}{
		{
			name: "no-limits-in-yaml",
			args: args{
				c: &apiv1.Container{
					Resources: apiv1.ResourceRequirements{},
				},
				r: model.ResourceRequirements{},
			},
			expectedRequests: map[apiv1.ResourceName]resource.Quantity{},
			expectedLimits:   map[apiv1.ResourceName]resource.Quantity{},
		},
		{
			name: "limits-in-yaml-no-limits-in-container",
			args: args{
				c: &apiv1.Container{
					Resources: apiv1.ResourceRequirements{},
				},
				r: model.ResourceRequirements{
					Limits: model.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("0.250Gi"),
						apiv1.ResourceCPU:    resource.MustParse("0.125"),
					},
					Requests: model.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("2Gi"),
						apiv1.ResourceCPU:    resource.MustParse("1"),
					},
				},
			},
			expectedRequests: map[apiv1.ResourceName]resource.Quantity{
				apiv1.ResourceMemory: resource.MustParse("2Gi"),
				apiv1.ResourceCPU:    resource.MustParse("1"),
			},
			expectedLimits: map[apiv1.ResourceName]resource.Quantity{
				apiv1.ResourceMemory: resource.MustParse("0.250Gi"),
				apiv1.ResourceCPU:    resource.MustParse("0.125"),
			},
		},
		{
			name: "no-limits-in-yaml-limits-in-container",
			args: args{
				c: &apiv1.Container{
					Resources: apiv1.ResourceRequirements{
						Limits: map[apiv1.ResourceName]resource.Quantity{
							apiv1.ResourceMemory: resource.MustParse("0.250Gi"),
							apiv1.ResourceCPU:    resource.MustParse("0.125"),
						},
						Requests: map[apiv1.ResourceName]resource.Quantity{
							apiv1.ResourceMemory: resource.MustParse("2Gi"),
							apiv1.ResourceCPU:    resource.MustParse("1"),
						},
					},
				},
				r: model.ResourceRequirements{},
			},
			expectedRequests: map[apiv1.ResourceName]resource.Quantity{
				apiv1.ResourceMemory: resource.MustParse("2Gi"),
				apiv1.ResourceCPU:    resource.MustParse("1"),
			},
			expectedLimits: map[apiv1.ResourceName]resource.Quantity{
				apiv1.ResourceMemory: resource.MustParse("0.250Gi"),
				apiv1.ResourceCPU:    resource.MustParse("0.125"),
			},
		},
		{
			name: "limits-in-yaml-limits-in-container",
			args: args{
				c: &apiv1.Container{
					Resources: apiv1.ResourceRequirements{
						Limits: map[apiv1.ResourceName]resource.Quantity{
							apiv1.ResourceMemory: resource.MustParse("0.250Gi"),
							apiv1.ResourceCPU:    resource.MustParse("0.125"),
						},
						Requests: map[apiv1.ResourceName]resource.Quantity{
							apiv1.ResourceMemory: resource.MustParse("2Gi"),
							apiv1.ResourceCPU:    resource.MustParse("1"),
						},
					},
				},
				r: model.ResourceRequirements{
					Limits: model.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("1Gi"),
						apiv1.ResourceCPU:    resource.MustParse("2"),
					},
					Requests: model.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("4Gi"),
						apiv1.ResourceCPU:    resource.MustParse("0.125"),
					},
				},
			},
			expectedRequests: map[apiv1.ResourceName]resource.Quantity{
				apiv1.ResourceMemory: resource.MustParse("4Gi"),
				apiv1.ResourceCPU:    resource.MustParse("0.125"),
			},
			expectedLimits: map[apiv1.ResourceName]resource.Quantity{
				apiv1.ResourceMemory: resource.MustParse("1Gi"),
				apiv1.ResourceCPU:    resource.MustParse("2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translateResources(tt.args.c, tt.args.r)

			a := tt.args.c.Resources.Requests[apiv1.ResourceMemory]
			b := tt.expectedRequests[apiv1.ResourceMemory]

			if a.Cmp(b) != 0 {
				t.Errorf("requests %s: expected %s, got %s", apiv1.ResourceMemory, b.String(), a.String())
			}

			a = tt.args.c.Resources.Requests[apiv1.ResourceCPU]
			b = tt.expectedRequests[apiv1.ResourceCPU]

			if a.Cmp(b) != 0 {
				t.Errorf("requests %s: expected %s, got %s", apiv1.ResourceCPU, b.String(), a.String())
			}

			a = tt.args.c.Resources.Limits[apiv1.ResourceMemory]
			b = tt.expectedLimits[apiv1.ResourceMemory]

			if a.Cmp(b) != 0 {
				t.Errorf("limits %s: expected %s, got %s", apiv1.ResourceMemory, b.String(), a.String())
			}

			a = tt.args.c.Resources.Limits[apiv1.ResourceCPU]
			b = tt.expectedLimits[apiv1.ResourceCPU]

			if a.Cmp(b) != 0 {
				t.Errorf("limits %s: expected %s, got %s", apiv1.ResourceCPU, b.String(), a.String())
			}
		})
	}
}

func Test_translateSecurityContext(t *testing.T) {
	tests := []struct {
		name         string
		c            *apiv1.Container
		s            *model.SecurityContext
		expectedAdd  []apiv1.Capability
		expectedDrop []apiv1.Capability
	}{
		{
			name: "single-add",
			c:    &apiv1.Container{},
			s: &model.SecurityContext{
				Capabilities: &model.Capabilities{
					Add: []apiv1.Capability{"SYS_TRACE"},
				},
			},
			expectedAdd: []apiv1.Capability{"SYS_TRACE"},
		},
		{
			name: "add-drop",
			c:    &apiv1.Container{},
			s: &model.SecurityContext{
				Capabilities: &model.Capabilities{
					Add:  []apiv1.Capability{"SYS_TRACE"},
					Drop: []apiv1.Capability{"SYS_NICE"},
				},
			},
			expectedAdd:  []apiv1.Capability{"SYS_TRACE"},
			expectedDrop: []apiv1.Capability{"SYS_NICE"},
		},
		{
			name: "merge-uniques",
			c: &apiv1.Container{
				SecurityContext: &apiv1.SecurityContext{
					Capabilities: &apiv1.Capabilities{
						Add:  []apiv1.Capability{"SYS_FOO"},
						Drop: []apiv1.Capability{"SYS_BAR"},
					},
				},
			},
			s: &model.SecurityContext{
				Capabilities: &model.Capabilities{
					Add:  []apiv1.Capability{"SYS_TRACE"},
					Drop: []apiv1.Capability{"SYS_NICE"},
				},
			},
			expectedAdd:  []apiv1.Capability{"SYS_FOO", "SYS_TRACE"},
			expectedDrop: []apiv1.Capability{"SYS_BAR", "SYS_NICE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translateSecurityContext(tt.c, tt.s)
			if tt.c.SecurityContext == nil {
				t.Fatal("SecurityContext was nil")
			}

			if !reflect.DeepEqual(tt.c.SecurityContext.Capabilities.Add, tt.expectedAdd) {
				t.Errorf("tt.c.SecurityContext.Capabilities.Add != tt.expectedAdd. Expected: %s, Got; %s", tt.expectedAdd, tt.c.SecurityContext.Capabilities.Add)
			}

			if !reflect.DeepEqual(tt.c.SecurityContext.Capabilities.Drop, tt.expectedDrop) {
				t.Errorf("tt.c.SecurityContext.Capabilities.Drop != tt.expectedDrop. Expected: %s, Got; %s", tt.expectedDrop, tt.c.SecurityContext.Capabilities.Drop)
			}
		})
	}
}
