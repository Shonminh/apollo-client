package agollo

// Config change types: add, modified, deleted
const (
	ADDED ConfigChangeType = iota
	MODIFIED
	DELETED
)

// ConfigChangeType specifies the type of a certain config
type ConfigChangeType int

// ChangeEvent is event of cofig change
type ChangeEvent struct {
	Namespace string
	Changes   []*ConfigChange
}

// ConfigChange contains config change info
type ConfigChange struct {
	Key        string
	OldValue   string
	NewValue   string
	ChangeType ConfigChangeType
}

func newModifyConfigChange(key string, oldValue string, newValue string) *ConfigChange {
	return &ConfigChange{
		Key:        key,
		OldValue:   oldValue,
		NewValue:   newValue,
		ChangeType: MODIFIED,
	}
}

func newAddConfigChange(key string, newValue string) *ConfigChange {
	return &ConfigChange{
		Key:        key,
		NewValue:   newValue,
		ChangeType: ADDED,
	}
}

func newDeletedConfigChange(key string, oldValue string) *ConfigChange {
	return &ConfigChange{
		Key:        key,
		OldValue:   oldValue,
		ChangeType: DELETED,
	}
}
