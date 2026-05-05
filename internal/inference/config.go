package inference

// InferenceConfig defines configurable behavior for schema inference.
type InferenceConfig struct {
	ArrayItemName string // default: "item"
}

func withDefaults(config InferenceConfig) InferenceConfig {
	if config.ArrayItemName == "" {
		config.ArrayItemName = "item"
	}
	return config
}
