package resources

import (
	"testing"

	"github.com/jenkinsci/kubernetes-operator/api/v1alpha2"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestGetJenkinsMasterPodBaseVolumes(t *testing.T) {
	t.Run("casc and groovy script with different configMap names", func(t *testing.T) {
		configMapName := "config-map"
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{
				ConfigurationAsCode: v1alpha2.ConfigurationAsCode{
					Customization: v1alpha2.Customization{
						Configurations: []v1alpha2.ConfigMapRef{
							{
								Name: configMapName,
							},
						},
						Secret: v1alpha2.SecretRef{
							Name: "casc-script",
						},
					},
				},
				GroovyScripts: v1alpha2.GroovyScripts{
					Customization: v1alpha2.Customization{
						Configurations: []v1alpha2.ConfigMapRef{
							{
								Name: configMapName,
							},
						},
						Secret: v1alpha2.SecretRef{
							Name: "groovy-script",
						},
					},
				},
			},
		}

		groovyExists, cascExists := checkSecretVolumesPresence(jenkins)

		assert.True(t, groovyExists)
		assert.True(t, cascExists)
	})
	t.Run("groovy script without secret name", func(t *testing.T) {
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{
				ConfigurationAsCode: v1alpha2.ConfigurationAsCode{
					Customization: v1alpha2.Customization{
						Configurations: []v1alpha2.ConfigMapRef{
							{
								Name: "casc-scripts",
							},
						},
						Secret: v1alpha2.SecretRef{
							Name: "jenkins-secret",
						},
					},
				},
				GroovyScripts: v1alpha2.GroovyScripts{
					Customization: v1alpha2.Customization{
						Configurations: []v1alpha2.ConfigMapRef{
							{
								Name: "groovy-scripts",
							},
						},
					},
				},
			},
		}

		groovyExists, cascExists := checkSecretVolumesPresence(jenkins)

		assert.True(t, cascExists)
		assert.False(t, groovyExists)
	})
	t.Run("casc without secret name", func(t *testing.T) {
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{
				ConfigurationAsCode: v1alpha2.ConfigurationAsCode{
					Customization: v1alpha2.Customization{
						Configurations: []v1alpha2.ConfigMapRef{
							{
								Name: "casc-scripts",
							},
						},
					},
				},
				GroovyScripts: v1alpha2.GroovyScripts{
					Customization: v1alpha2.Customization{
						Configurations: []v1alpha2.ConfigMapRef{
							{
								Name: "groovy-scripts",
							},
						},
						Secret: v1alpha2.SecretRef{
							Name: "jenkins-secret",
						},
					},
				},
			},
		}

		groovyExists, cascExists := checkSecretVolumesPresence(jenkins)

		assert.True(t, groovyExists)
		assert.False(t, cascExists)
	})
	t.Run("casc and groovy script shared secret name", func(t *testing.T) {
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{
				ConfigurationAsCode: v1alpha2.ConfigurationAsCode{
					Customization: v1alpha2.Customization{
						Configurations: []v1alpha2.ConfigMapRef{
							{
								Name: "casc-scripts",
							},
						},
						Secret: v1alpha2.SecretRef{
							Name: "jenkins-secret",
						},
					},
				},
				GroovyScripts: v1alpha2.GroovyScripts{
					Customization: v1alpha2.Customization{
						Configurations: []v1alpha2.ConfigMapRef{
							{
								Name: "groovy-scripts",
							},
						},
						Secret: v1alpha2.SecretRef{
							Name: "jenkins-secret",
						},
					},
				},
			},
		}

		groovyExists, cascExists := checkSecretVolumesPresence(jenkins)

		assert.True(t, groovyExists)
		assert.True(t, cascExists)
	})
	t.Run("empty jenkins spec adds default jenkins-home volume", func(t *testing.T) {
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{},
		}

		volumes := GetJenkinsMasterPodBaseVolumes(jenkins)

		// Should have jenkins-home volume as first element
		found := false
		for _, volume := range volumes {
			if volume.Name == JenkinsHomeVolumeName {
				found = true
				assert.NotNil(t, volume.VolumeSource.EmptyDir)
				break
			}
		}
		assert.True(t, found, "jenkins-home volume not found")
	})
	t.Run("user-defined jenkins-home volume should not add default", func(t *testing.T) {
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{
				Master: v1alpha2.JenkinsMaster{
					Volumes: []corev1.Volume{
						{
							Name: JenkinsHomeVolumeName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "my-pvc",
								},
							},
						},
					},
				},
			},
		}

		volumes := GetJenkinsMasterPodBaseVolumes(jenkins)

		// Should NOT have jenkins-home volume in base volumes (user provides it)
		jenkinsHomeCount := 0
		for _, volume := range volumes {
			if volume.Name == JenkinsHomeVolumeName {
				jenkinsHomeCount++
			}
		}
		assert.Equal(t, 0, jenkinsHomeCount, "jenkins-home volume should not be in base volumes when user defines it")
	})
	t.Run("user-defined other volume should still add default jenkins-home", func(t *testing.T) {
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{
				Master: v1alpha2.JenkinsMaster{
					Volumes: []corev1.Volume{
						{
							Name: "some-other-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		}

		volumes := GetJenkinsMasterPodBaseVolumes(jenkins)

		// Should have jenkins-home volume
		found := false
		for _, volume := range volumes {
			if volume.Name == JenkinsHomeVolumeName {
				found = true
				assert.NotNil(t, volume.VolumeSource.EmptyDir)
				break
			}
		}
		assert.True(t, found, "jenkins-home volume not found")
	})
}

func checkSecretVolumesPresence(jenkins *v1alpha2.Jenkins) (groovyExists bool, cascExists bool) {
	for _, volume := range GetJenkinsMasterPodBaseVolumes(jenkins) {
		if volume.Name == ("gs-" + jenkins.Spec.GroovyScripts.Secret.Name) {
			groovyExists = true
		} else if volume.Name == ("casc-" + jenkins.Spec.ConfigurationAsCode.Secret.Name) {
			cascExists = true
		}
	}
	return groovyExists, cascExists
}

func TestGetJenkinsMasterContainerBaseVolumeMounts(t *testing.T) {
	t.Run("empty containers", func(t *testing.T) {
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{
				Master: v1alpha2.JenkinsMaster{
					Containers: []v1alpha2.Container{},
				},
			},
		}

		volumeMounts := GetJenkinsMasterContainerBaseVolumeMounts(jenkins)

		// Should have base volume mounts (scripts, init-configuration, operator-credentials)
		assert.Len(t, volumeMounts, 3)
		assert.Equal(t, jenkinsScriptsVolumeName, volumeMounts[0].Name)
		assert.Equal(t, jenkinsInitConfigurationVolumeName, volumeMounts[1].Name)
		assert.Equal(t, jenkinsOperatorCredentialsVolumeName, volumeMounts[2].Name)
	})

	t.Run("container without jenkins-home volume mount", func(t *testing.T) {
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{
				Master: v1alpha2.JenkinsMaster{
					Containers: []v1alpha2.Container{
						{
							Name: "jenkins-master",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "some-other-volume",
									MountPath: "/some/path",
								},
							},
						},
					},
				},
			},
		}

		volumeMounts := GetJenkinsMasterContainerBaseVolumeMounts(jenkins)

		// Should have base volume mounts only (no jenkins-home)
		assert.Len(t, volumeMounts, 3)
		assert.Equal(t, jenkinsScriptsVolumeName, volumeMounts[0].Name)
	})

	t.Run("container with jenkins-home volume mount", func(t *testing.T) {
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{
				Master: v1alpha2.JenkinsMaster{
					Containers: []v1alpha2.Container{
						{
							Name: "jenkins-master",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      JenkinsHomeVolumeName,
									MountPath: "/var/lib/jenkins",
								},
							},
						},
					},
				},
			},
		}

		volumeMounts := GetJenkinsMasterContainerBaseVolumeMounts(jenkins)

		// Should have jenkins-home as first element + base volume mounts
		assert.Len(t, volumeMounts, 4)
		assert.Equal(t, JenkinsHomeVolumeName, volumeMounts[0].Name)
		assert.Equal(t, "/var/lib/jenkins", volumeMounts[0].MountPath)
		assert.Equal(t, jenkinsScriptsVolumeName, volumeMounts[1].Name)
	})

	t.Run("container with jenkins-home volume mount with custom path", func(t *testing.T) {
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{
				Master: v1alpha2.JenkinsMaster{
					Containers: []v1alpha2.Container{
						{
							Name: "jenkins-master",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      JenkinsHomeVolumeName,
									MountPath: "/custom/jenkins/home",
								},
							},
						},
					},
				},
			},
		}

		volumeMounts := GetJenkinsMasterContainerBaseVolumeMounts(jenkins)

		// Should have jenkins-home as first element with custom path
		assert.Len(t, volumeMounts, 4)
		assert.Equal(t, JenkinsHomeVolumeName, volumeMounts[0].Name)
		assert.Equal(t, "/custom/jenkins/home", volumeMounts[0].MountPath)
	})

	t.Run("with groovy scripts secret", func(t *testing.T) {
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{
				Master: v1alpha2.JenkinsMaster{
					Containers: []v1alpha2.Container{},
				},
				GroovyScripts: v1alpha2.GroovyScripts{
					Customization: v1alpha2.Customization{
						Secret: v1alpha2.SecretRef{
							Name: "groovy-secret",
						},
					},
				},
			},
		}

		volumeMounts := GetJenkinsMasterContainerBaseVolumeMounts(jenkins)

		// Should have base volume mounts + groovy scripts secret
		assert.Len(t, volumeMounts, 4)
		found := false
		for _, vm := range volumeMounts {
			if vm.Name == "gs-groovy-secret" {
				found = true
				assert.Equal(t, GroovyScriptsSecretVolumePath, vm.MountPath)
				assert.True(t, vm.ReadOnly)
			}
		}
		assert.True(t, found, "groovy scripts secret volume mount not found")
	})

	t.Run("with configuration as code secret", func(t *testing.T) {
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{
				Master: v1alpha2.JenkinsMaster{
					Containers: []v1alpha2.Container{},
				},
				ConfigurationAsCode: v1alpha2.ConfigurationAsCode{
					Customization: v1alpha2.Customization{
						Secret: v1alpha2.SecretRef{
							Name: "casc-secret",
						},
					},
				},
			},
		}

		volumeMounts := GetJenkinsMasterContainerBaseVolumeMounts(jenkins)

		// Should have base volume mounts + casc secret
		assert.Len(t, volumeMounts, 4)
		found := false
		for _, vm := range volumeMounts {
			if vm.Name == "casc-casc-secret" {
				found = true
				assert.Equal(t, ConfigurationAsCodeSecretVolumePath, vm.MountPath)
				assert.True(t, vm.ReadOnly)
			}
		}
		assert.True(t, found, "casc secret volume mount not found")
	})

	t.Run("with jenkins-home and both secrets", func(t *testing.T) {
		jenkins := &v1alpha2.Jenkins{
			Spec: v1alpha2.JenkinsSpec{
				Master: v1alpha2.JenkinsMaster{
					Containers: []v1alpha2.Container{
						{
							Name: "jenkins-master",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      JenkinsHomeVolumeName,
									MountPath: "/var/lib/jenkins",
								},
							},
						},
					},
				},
				GroovyScripts: v1alpha2.GroovyScripts{
					Customization: v1alpha2.Customization{
						Secret: v1alpha2.SecretRef{
							Name: "groovy-secret",
						},
					},
				},
				ConfigurationAsCode: v1alpha2.ConfigurationAsCode{
					Customization: v1alpha2.Customization{
						Secret: v1alpha2.SecretRef{
							Name: "casc-secret",
						},
					},
				},
			},
		}

		volumeMounts := GetJenkinsMasterContainerBaseVolumeMounts(jenkins)

		// Should have jenkins-home + base volume mounts + both secrets
		assert.Len(t, volumeMounts, 6)
		assert.Equal(t, JenkinsHomeVolumeName, volumeMounts[0].Name)
	})
}
