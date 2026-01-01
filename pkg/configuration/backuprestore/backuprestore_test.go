package backuprestore

import (
	"testing"
	"time"

	"github.com/jenkinsci/kubernetes-operator/api/v1alpha2"
	"github.com/jenkinsci/kubernetes-operator/pkg/configuration"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	testBackupContainerName = "backup"
)

func getTestLogger() logr.Logger {
	return zap.New(zap.UseDevMode(true))
}

func getBaseJenkins() *v1alpha2.Jenkins {
	return &v1alpha2.Jenkins{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-jenkins",
			Namespace: "default",
		},
		Spec: v1alpha2.JenkinsSpec{
			Master: v1alpha2.JenkinsMaster{
				Containers: []v1alpha2.Container{
					{Name: "jenkins-master"},
					{Name: testBackupContainerName},
				},
			},
		},
	}
}

func TestNew(t *testing.T) {
	t.Run("creates new BackupAndRestore instance", func(t *testing.T) {
		jenkins := getBaseJenkins()
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		logger := getTestLogger()

		bar := New(config, logger)

		assert.NotNil(t, bar)
		assert.Equal(t, config.Jenkins, bar.Configuration.Jenkins)
	})
}

func TestBackupAndRestore_Validate(t *testing.T) {
	t.Run("no errors when backup and restore are not configured", func(t *testing.T) {
		jenkins := getBaseJenkins()
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		messages := bar.Validate()

		assert.Empty(t, messages)
	})

	t.Run("error when restore container not found", func(t *testing.T) {
		jenkins := getBaseJenkins()
		jenkins.Spec.Restore.ContainerName = "non-existent-container"
		jenkins.Spec.Restore.Action.Exec = &corev1.ExecAction{Command: []string{"restore.sh"}}
		jenkins.Spec.Backup.ContainerName = testBackupContainerName
		jenkins.Spec.Backup.Action.Exec = &corev1.ExecAction{Command: []string{"backup.sh"}}
		jenkins.Spec.Backup.Interval = 30
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		messages := bar.Validate()

		assert.Contains(t, messages, "restore container 'non-existent-container' not found in CR spec.master.containers")
	})

	t.Run("error when backup container not found", func(t *testing.T) {
		jenkins := getBaseJenkins()
		jenkins.Spec.Backup.ContainerName = "non-existent-container"
		jenkins.Spec.Backup.Action.Exec = &corev1.ExecAction{Command: []string{"backup.sh"}}
		jenkins.Spec.Backup.Interval = 30
		jenkins.Spec.Restore.ContainerName = testBackupContainerName
		jenkins.Spec.Restore.Action.Exec = &corev1.ExecAction{Command: []string{"restore.sh"}}
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		messages := bar.Validate()

		assert.Contains(t, messages, "backup container 'non-existent-container' not found in CR spec.master.containers")
	})

	t.Run("error when restore action exec is not configured", func(t *testing.T) {
		jenkins := getBaseJenkins()
		jenkins.Spec.Restore.ContainerName = testBackupContainerName
		// Action.Exec is nil
		jenkins.Spec.Backup.ContainerName = testBackupContainerName
		jenkins.Spec.Backup.Action.Exec = &corev1.ExecAction{Command: []string{"backup.sh"}}
		jenkins.Spec.Backup.Interval = 30
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		messages := bar.Validate()

		assert.Contains(t, messages, "spec.restore.action.exec is not configured")
	})

	t.Run("error when backup action exec is not configured", func(t *testing.T) {
		jenkins := getBaseJenkins()
		jenkins.Spec.Backup.ContainerName = testBackupContainerName
		// Action.Exec is nil
		jenkins.Spec.Backup.Interval = 30
		jenkins.Spec.Restore.ContainerName = testBackupContainerName
		jenkins.Spec.Restore.Action.Exec = &corev1.ExecAction{Command: []string{"restore.sh"}}
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		messages := bar.Validate()

		assert.Contains(t, messages, "spec.backup.action.exec is not configured")
	})

	t.Run("error when backup interval is not configured", func(t *testing.T) {
		jenkins := getBaseJenkins()
		jenkins.Spec.Backup.ContainerName = testBackupContainerName
		jenkins.Spec.Backup.Action.Exec = &corev1.ExecAction{Command: []string{"backup.sh"}}
		jenkins.Spec.Backup.Interval = 0
		jenkins.Spec.Restore.ContainerName = testBackupContainerName
		jenkins.Spec.Restore.Action.Exec = &corev1.ExecAction{Command: []string{"restore.sh"}}
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		messages := bar.Validate()

		assert.Contains(t, messages, "spec.backup.interval is not configured")
	})

	t.Run("error when restore is configured but backup is not", func(t *testing.T) {
		jenkins := getBaseJenkins()
		jenkins.Spec.Restore.ContainerName = testBackupContainerName
		jenkins.Spec.Restore.Action.Exec = &corev1.ExecAction{Command: []string{"restore.sh"}}
		// Backup is not configured
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		messages := bar.Validate()

		assert.Contains(t, messages, "spec.backup.containerName is not configured")
	})

	t.Run("error when backup is configured but restore is not", func(t *testing.T) {
		jenkins := getBaseJenkins()
		jenkins.Spec.Backup.ContainerName = testBackupContainerName
		jenkins.Spec.Backup.Action.Exec = &corev1.ExecAction{Command: []string{"backup.sh"}}
		jenkins.Spec.Backup.Interval = 30
		// Restore is not configured
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		messages := bar.Validate()

		assert.Contains(t, messages, "spec.restore.containerName is not configured")
	})

	t.Run("valid configuration with all required fields", func(t *testing.T) {
		jenkins := getBaseJenkins()
		jenkins.Spec.Backup.ContainerName = testBackupContainerName
		jenkins.Spec.Backup.Action.Exec = &corev1.ExecAction{Command: []string{"backup.sh"}}
		jenkins.Spec.Backup.Interval = 30
		jenkins.Spec.Restore.ContainerName = testBackupContainerName
		jenkins.Spec.Restore.Action.Exec = &corev1.ExecAction{Command: []string{"restore.sh"}}
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		messages := bar.Validate()

		assert.Empty(t, messages)
	})
}

func TestBackupTriggers_Key(t *testing.T) {
	t.Run("generates correct key", func(t *testing.T) {
		bt := &backupTriggers{triggers: make(map[string]backupTrigger)}

		key := bt.key("test-namespace", "test-name")

		assert.Equal(t, "test-namespace/test-name", key)
	})
}

func TestBackupTriggers_AddGetStop(t *testing.T) {
	// Reset global triggers before test
	triggers = backupTriggers{triggers: make(map[string]backupTrigger)}

	t.Run("add and get trigger", func(t *testing.T) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		trigger := backupTrigger{
			interval: 30,
			ticker:   ticker,
		}

		triggers.add("ns1", "name1", trigger)

		retrieved, found := triggers.get("ns1", "name1")
		assert.True(t, found)
		assert.Equal(t, uint64(30), retrieved.interval)
	})

	t.Run("get non-existent trigger returns false", func(t *testing.T) {
		_, found := triggers.get("non-existent", "trigger")
		assert.False(t, found)
	})

	t.Run("stop trigger removes it", func(t *testing.T) {
		ticker := time.NewTicker(1 * time.Second)
		trigger := backupTrigger{
			interval: 60,
			ticker:   ticker,
		}
		triggers.add("ns2", "name2", trigger)

		triggers.stop(getTestLogger(), "ns2", "name2")

		_, found := triggers.get("ns2", "name2")
		assert.False(t, found)
	})

	t.Run("stop non-existent trigger does not panic", func(t *testing.T) {
		// Should not panic
		triggers.stop(getTestLogger(), "non-existent", "trigger")
	})
}

func TestBackupAndRestore_IsBackupTriggerEnabled(t *testing.T) {
	// Reset global triggers before test
	triggers = backupTriggers{triggers: make(map[string]backupTrigger)}

	t.Run("returns false when no trigger exists", func(t *testing.T) {
		jenkins := getBaseJenkins()
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		enabled := bar.IsBackupTriggerEnabled()

		assert.False(t, enabled)
	})

	t.Run("returns true when trigger exists", func(t *testing.T) {
		jenkins := getBaseJenkins()
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		// Add a trigger
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		triggers.add(jenkins.Namespace, jenkins.Name, backupTrigger{
			interval: 30,
			ticker:   ticker,
		})

		enabled := bar.IsBackupTriggerEnabled()

		assert.True(t, enabled)

		// Clean up
		triggers.stop(getTestLogger(), jenkins.Namespace, jenkins.Name)
	})
}

func TestBackupAndRestore_StopBackupTrigger(t *testing.T) {
	// Reset global triggers before test
	triggers = backupTriggers{triggers: make(map[string]backupTrigger)}

	t.Run("stops existing trigger", func(t *testing.T) {
		jenkins := getBaseJenkins()
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		// Add a trigger
		ticker := time.NewTicker(1 * time.Second)
		triggers.add(jenkins.Namespace, jenkins.Name, backupTrigger{
			interval: 30,
			ticker:   ticker,
		})

		bar.StopBackupTrigger()

		_, found := triggers.get(jenkins.Namespace, jenkins.Name)
		assert.False(t, found)
	})

	t.Run("does not panic when no trigger exists", func(t *testing.T) {
		jenkins := getBaseJenkins()
		jenkins.Name = "non-existent-jenkins"
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		// Should not panic
		bar.StopBackupTrigger()
	})
}

func TestBackupAndRestore_EnsureBackupTrigger(t *testing.T) {
	t.Run("does nothing when backup is not configured", func(t *testing.T) {
		// Reset global triggers before test
		triggers = backupTriggers{triggers: make(map[string]backupTrigger)}

		jenkins := getBaseJenkins()
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		err := bar.EnsureBackupTrigger()

		assert.NoError(t, err)
		assert.False(t, bar.IsBackupTriggerEnabled())
	})

	t.Run("starts trigger when backup is configured and no trigger exists", func(t *testing.T) {
		// Reset global triggers before test
		triggers = backupTriggers{triggers: make(map[string]backupTrigger)}

		jenkins := getBaseJenkins()
		jenkins.Spec.Backup.ContainerName = testBackupContainerName
		jenkins.Spec.Backup.Interval = 30
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		err := bar.EnsureBackupTrigger()

		assert.NoError(t, err)
		assert.True(t, bar.IsBackupTriggerEnabled())

		// Clean up
		bar.StopBackupTrigger()
	})

	t.Run("stops trigger when backup config is removed", func(t *testing.T) {
		// Reset global triggers before test
		triggers = backupTriggers{triggers: make(map[string]backupTrigger)}

		jenkins := getBaseJenkins()
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		// Add a trigger manually
		ticker := time.NewTicker(1 * time.Second)
		triggers.add(jenkins.Namespace, jenkins.Name, backupTrigger{
			interval: 30,
			ticker:   ticker,
		})

		// Now ensure trigger - should stop it since backup is not configured
		err := bar.EnsureBackupTrigger()

		assert.NoError(t, err)
		assert.False(t, bar.IsBackupTriggerEnabled())
	})

	t.Run("restarts trigger when interval changes", func(t *testing.T) {
		// Reset global triggers before test
		triggers = backupTriggers{triggers: make(map[string]backupTrigger)}

		jenkins := getBaseJenkins()
		jenkins.Spec.Backup.ContainerName = testBackupContainerName
		jenkins.Spec.Backup.Interval = 60
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		// Add a trigger with different interval
		ticker := time.NewTicker(1 * time.Second)
		triggers.add(jenkins.Namespace, jenkins.Name, backupTrigger{
			interval: 30, // Different from 60
			ticker:   ticker,
		})

		err := bar.EnsureBackupTrigger()

		assert.NoError(t, err)
		assert.True(t, bar.IsBackupTriggerEnabled())

		trigger, found := triggers.get(jenkins.Namespace, jenkins.Name)
		assert.True(t, found)
		assert.Equal(t, uint64(60), trigger.interval)

		// Clean up
		bar.StopBackupTrigger()
	})

	t.Run("does nothing when interval is the same", func(t *testing.T) {
		// Reset global triggers before test
		triggers = backupTriggers{triggers: make(map[string]backupTrigger)}

		jenkins := getBaseJenkins()
		jenkins.Spec.Backup.ContainerName = testBackupContainerName
		jenkins.Spec.Backup.Interval = 30
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		// Add a trigger with same interval
		ticker := time.NewTicker(1 * time.Second)
		triggers.add(jenkins.Namespace, jenkins.Name, backupTrigger{
			interval: 30,
			ticker:   ticker,
		})

		err := bar.EnsureBackupTrigger()

		assert.NoError(t, err)
		assert.True(t, bar.IsBackupTriggerEnabled())

		// Clean up
		bar.StopBackupTrigger()
	})
}

func TestBackupAndRestore_Restore(t *testing.T) {
	t.Run("skips restore when restore container not configured", func(t *testing.T) {
		jenkins := getBaseJenkins()
		// Restore.ContainerName is empty
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		err := bar.Restore(nil)

		assert.NoError(t, err)
	})

	t.Run("skips restore when restore action exec is nil", func(t *testing.T) {
		jenkins := getBaseJenkins()
		jenkins.Spec.Restore.ContainerName = testBackupContainerName
		// Action.Exec is nil
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		err := bar.Restore(nil)

		assert.NoError(t, err)
	})

	t.Run("skips restore when already restored", func(t *testing.T) {
		jenkins := getBaseJenkins()
		jenkins.Spec.Restore.ContainerName = testBackupContainerName
		jenkins.Spec.Restore.Action.Exec = &corev1.ExecAction{Command: []string{"restore.sh"}}
		jenkins.Status.RestoredBackup = 1 // Already restored
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		err := bar.Restore(nil)

		assert.NoError(t, err)
	})
}

func TestBackupAndRestore_Backup(t *testing.T) {
	t.Run("skips backup when backup container not configured", func(t *testing.T) {
		jenkins := getBaseJenkins()
		// Backup.ContainerName is empty
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		err := bar.Backup(false)

		assert.NoError(t, err)
	})

	t.Run("skips backup when backup action exec is nil", func(t *testing.T) {
		jenkins := getBaseJenkins()
		jenkins.Spec.Backup.ContainerName = testBackupContainerName
		// Action.Exec is nil
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		err := bar.Backup(false)

		assert.NoError(t, err)
	})

	t.Run("skips backup when pending equals last backup", func(t *testing.T) {
		jenkins := getBaseJenkins()
		jenkins.Spec.Backup.ContainerName = testBackupContainerName
		jenkins.Spec.Backup.Action.Exec = &corev1.ExecAction{Command: []string{"backup.sh"}}
		jenkins.Status.PendingBackup = 5
		jenkins.Status.LastBackup = 5 // Same as pending
		config := configuration.Configuration{
			Jenkins: jenkins,
		}
		bar := New(config, getTestLogger())

		err := bar.Backup(false)

		assert.NoError(t, err)
	})
}

func TestNoBackupConstant(t *testing.T) {
	t.Run("noBackup constant is -1", func(t *testing.T) {
		assert.Equal(t, "-1", noBackup)
	})
}
