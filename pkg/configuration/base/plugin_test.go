package base

import (
	"testing"

	"github.com/bndr/gojenkins"
	"github.com/stretchr/testify/assert"

	"github.com/jenkinsci/kubernetes-operator/api/v1alpha2"
)

func Test_isPluginVersionCompatible(t *testing.T) {
	t.Run("compatible version", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "my-plugin",
						Version:   "1.0.0",
						Active:    true,
						Enabled:   true,
						Deleted:   false,
					},
				},
			},
		}
		plugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0",
		}

		got, ok := isPluginVersionCompatible(plugins, plugin)

		assert.True(t, ok)
		assert.Equal(t, "my-plugin", got.ShortName)
		assert.Equal(t, "1.0.0", got.Version)
	})
	t.Run("incompatible version", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "my-plugin",
						Version:   "2.0.0",
						Active:    true,
						Enabled:   true,
						Deleted:   false,
					},
				},
			},
		}
		plugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0",
		}

		got, ok := isPluginVersionCompatible(plugins, plugin)

		assert.False(t, ok)
		assert.Equal(t, "my-plugin", got.ShortName)
		assert.Equal(t, "2.0.0", got.Version)
	})
	t.Run("plugin version is latest returns false", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "my-plugin",
						Version:   "1.0.0",
						Active:    true,
						Enabled:   true,
						Deleted:   false,
					},
				},
			},
		}
		plugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "latest",
		}

		got, ok := isPluginVersionCompatible(plugins, plugin)

		assert.False(t, ok)
		assert.Equal(t, "", got.ShortName)
		assert.Equal(t, "", got.Version)
	})
	t.Run("plugin not found", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "other-plugin",
						Version:   "1.0.0",
						Active:    true,
						Enabled:   true,
						Deleted:   false,
					},
				},
			},
		}
		plugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0",
		}

		got, ok := isPluginVersionCompatible(plugins, plugin)

		assert.False(t, ok)
		assert.Equal(t, "", got.ShortName)
	})
	t.Run("plugin with empty version in jenkins", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "my-plugin",
						Version:   "",
						Active:    true,
						Enabled:   true,
						Deleted:   false,
					},
				},
			},
		}
		plugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0",
		}

		got, ok := isPluginVersionCompatible(plugins, plugin)

		assert.False(t, ok)
		assert.Equal(t, "", got.ShortName)
	})
	t.Run("empty plugins list", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{},
			},
		}
		plugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0",
		}

		got, ok := isPluginVersionCompatible(plugins, plugin)

		assert.False(t, ok)
		assert.Equal(t, "", got.ShortName)
	})
	t.Run("multiple plugins with matching one", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "plugin-a",
						Version:   "1.0.0",
						Active:    true,
						Enabled:   true,
						Deleted:   false,
					},
					{
						ShortName: "my-plugin",
						Version:   "2.5.0",
						Active:    true,
						Enabled:   true,
						Deleted:   false,
					},
					{
						ShortName: "plugin-b",
						Version:   "3.0.0",
						Active:    true,
						Enabled:   true,
						Deleted:   false,
					},
				},
			},
		}
		plugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "2.5.0",
		}

		got, ok := isPluginVersionCompatible(plugins, plugin)

		assert.True(t, ok)
		assert.Equal(t, "my-plugin", got.ShortName)
		assert.Equal(t, "2.5.0", got.Version)
	})
	t.Run("semantic version comparison is exact", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "my-plugin",
						Version:   "1.0.1",
						Active:    true,
						Enabled:   true,
						Deleted:   false,
					},
				},
			},
		}
		plugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0",
		}

		got, ok := isPluginVersionCompatible(plugins, plugin)

		assert.False(t, ok)
		assert.Equal(t, "my-plugin", got.ShortName)
		assert.Equal(t, "1.0.1", got.Version)
	})
	t.Run("plugin with suffix version", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "my-plugin",
						Version:   "1.0.0-SNAPSHOT",
						Active:    true,
						Enabled:   true,
						Deleted:   false,
					},
				},
			},
		}
		plugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0-SNAPSHOT",
		}

		got, ok := isPluginVersionCompatible(plugins, plugin)

		assert.True(t, ok)
		assert.Equal(t, "my-plugin", got.ShortName)
		assert.Equal(t, "1.0.0-SNAPSHOT", got.Version)
	})
	t.Run("plugin with suffix version mismatch", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "my-plugin",
						Version:   "1.0.0-SNAPSHOT",
						Active:    true,
						Enabled:   true,
						Deleted:   false,
					},
				},
			},
		}
		plugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0",
		}

		got, ok := isPluginVersionCompatible(plugins, plugin)

		assert.False(t, ok)
		assert.Equal(t, "my-plugin", got.ShortName)
		assert.Equal(t, "1.0.0-SNAPSHOT", got.Version)
	})
}

func Test_isValidPlugin(t *testing.T) {
	t.Run("valid plugin", func(t *testing.T) {
		plugin := gojenkins.Plugin{
			Active:  true,
			Enabled: true,
			Deleted: false,
		}

		got := isValidPlugin(plugin)

		assert.True(t, got)
	})
	t.Run("inactive plugin", func(t *testing.T) {
		plugin := gojenkins.Plugin{
			Active:  false,
			Enabled: true,
			Deleted: false,
		}

		got := isValidPlugin(plugin)

		assert.False(t, got)
	})
	t.Run("disabled plugin", func(t *testing.T) {
		plugin := gojenkins.Plugin{
			Active:  true,
			Enabled: false,
			Deleted: false,
		}

		got := isValidPlugin(plugin)

		assert.False(t, got)
	})
	t.Run("deleted plugin", func(t *testing.T) {
		plugin := gojenkins.Plugin{
			Active:  true,
			Enabled: true,
			Deleted: true,
		}

		got := isValidPlugin(plugin)

		assert.False(t, got)
	})
	t.Run("all flags false", func(t *testing.T) {
		plugin := gojenkins.Plugin{
			Active:  false,
			Enabled: false,
			Deleted: false,
		}

		got := isValidPlugin(plugin)

		assert.False(t, got)
	})
}

func Test_isPluginInstalled(t *testing.T) {
	t.Run("plugin installed and valid", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "my-plugin",
						Version:   "1.0.0",
						Active:    true,
						Enabled:   true,
						Deleted:   false,
					},
				},
			},
		}
		requiredPlugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0",
		}

		got, ok := isPluginInstalled(plugins, requiredPlugin)

		assert.True(t, ok)
		assert.Equal(t, "my-plugin", got.ShortName)
	})
	t.Run("plugin installed but inactive", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "my-plugin",
						Version:   "1.0.0",
						Active:    false,
						Enabled:   true,
						Deleted:   false,
					},
				},
			},
		}
		requiredPlugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0",
		}

		got, ok := isPluginInstalled(plugins, requiredPlugin)

		assert.False(t, ok)
		assert.Equal(t, "my-plugin", got.ShortName)
	})
	t.Run("plugin installed but disabled", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "my-plugin",
						Version:   "1.0.0",
						Active:    true,
						Enabled:   false,
						Deleted:   false,
					},
				},
			},
		}
		requiredPlugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0",
		}

		got, ok := isPluginInstalled(plugins, requiredPlugin)

		assert.False(t, ok)
		assert.Equal(t, "my-plugin", got.ShortName)
	})
	t.Run("plugin installed but deleted", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "my-plugin",
						Version:   "1.0.0",
						Active:    true,
						Enabled:   true,
						Deleted:   true,
					},
				},
			},
		}
		requiredPlugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0",
		}

		got, ok := isPluginInstalled(plugins, requiredPlugin)

		assert.False(t, ok)
		assert.Equal(t, "my-plugin", got.ShortName)
	})
	t.Run("plugin not found", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{
					{
						ShortName: "other-plugin",
						Version:   "1.0.0",
						Active:    true,
						Enabled:   true,
						Deleted:   false,
					},
				},
			},
		}
		requiredPlugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0",
		}

		got, ok := isPluginInstalled(plugins, requiredPlugin)

		assert.False(t, ok)
		assert.Equal(t, "", got.ShortName)
	})
	t.Run("empty plugins list", func(t *testing.T) {
		plugins := &gojenkins.Plugins{
			Raw: &gojenkins.PluginResponse{
				Plugins: []gojenkins.Plugin{},
			},
		}
		requiredPlugin := v1alpha2.Plugin{
			Name:    "my-plugin",
			Version: "1.0.0",
		}

		got, ok := isPluginInstalled(plugins, requiredPlugin)

		assert.False(t, ok)
		assert.Equal(t, "", got.ShortName)
	})
}
