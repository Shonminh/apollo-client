package agollo

import (
	"fmt"
	"github.com/Shonminh/apollo-client/internal/apollo"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestCache_getChangeEvent(t *testing.T) {
	gOption = newDefaultOption()
	initCache(2, false)
	nl := []*namespace{
		{
			NamespaceName: "immutable",
			releaseKey:    "immutable",
		},
		{
			NamespaceName: "mutable",
			releaseKey:    "mutable",
		},
	}

	configSets := []map[string]string{
		{
			"imkey1": "imvalue1",
			"imkey2": "imvalue2",
		},
		{
			"key1": "value1",
			"key2": "value2",
		},
	}

	/*firstExpectChanges := []*ChangeEvent{
		{
			Namespace: "immutable",
			Changes: []*ConfigChange{
				{
					Key:        "imkey1",
					NewValue:   "imvalue1",
					ChangeType: 0,
				},
				{
					Key:        "imkey2",
					NewValue:   "imvalue2",
					ChangeType: 0,
				},
			},
		},
		{
			Namespace: "mutable",
			Changes: []*ConfigChange{
				{
					Key:        "key1",
					NewValue:   "value1",
					ChangeType: 0,
				},
				{
					Key:        "key2",
					NewValue:   "value2",
					ChangeType: 0,
				},
			},
		},
	}*/

	configChangeSets := []map[string]string{
		{
			"imkey1": "im modify1",
			"imkey3": "im add3",
		},
		{
			"key1": "modify1",
			"key3": "add3",
		},
	}
	secondExpectChanges := []*ChangeEvent{
		{
			Namespace: "immutable",
			Changes: []*ConfigChange{
				{
					Key:        "imkey3",
					NewValue:   "im add3",
					ChangeType: 0,
				},
				{
					Key:        "imkey1",
					OldValue:   "imvalue1",
					NewValue:   "im modify1",
					ChangeType: 1,
				},
				{
					Key:        "imkey2",
					OldValue:   "imvalue2",
					ChangeType: 2,
				},
			},
		},
		{
			Namespace: "mutable",
			Changes: []*ConfigChange{
				{
					Key:        "key3",
					NewValue:   "add3",
					ChangeType: 0,
				},
				{
					Key:        "key1",
					OldValue:   "value1",
					NewValue:   "modify1",
					ChangeType: 1,
				},
				{
					Key:        "key2",
					OldValue:   "value2",
					ChangeType: 2,
				},
			},
		},
	}

	for i, ns := range nl {
		cfg := &apollo.Config{
			ConnConfig:     apollo.ConnConfig{NamespaceName: ns.NamespaceName},
			Configurations: configSets[i],
		}
		changeEvent := getChangeEvent(cfg)
		updateCache(cfg, ns, changeEvent)

		//assert.Equal(t, firstExpectChanges[i], changeEvent)
		fmt.Printf("nameSpace: %+v\n", changeEvent.Namespace)
		for i := range changeEvent.Changes {
			fmt.Printf("%+v\n", changeEvent.Changes[i])
		}
	}
	b, _ := gConfigCache.Get([]byte("mutable.key1"))
	fmt.Printf("%v\n", string(b))

	for i, ns := range nl {
		cfg := &apollo.Config{
			ConnConfig:     apollo.ConnConfig{NamespaceName: ns.NamespaceName},
			Configurations: configChangeSets[i],
		}
		cl := getChangeEvent(cfg)
		updateCache(cfg, ns, cl)

		assert.Equal(t, secondExpectChanges[i], sortChanges(cl))
		fmt.Printf("nameSpace: %+v\n", cl.Namespace)
		for i := range cl.Changes {
			fmt.Printf("%+v\n", cl.Changes[i])
		}
	}
}

type sConfigChange []*ConfigChange

func (s sConfigChange) Len() int {
	return len(s)
}

func (s sConfigChange) Less(i, j int) bool {
	return s[i].ChangeType < s[j].ChangeType
}

func (s sConfigChange) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func sortChanges(changes *ChangeEvent) *ChangeEvent {
	sort.Sort(sConfigChange(changes.Changes))
	return changes
}

func TestCache_getChangeEvent_ignore(t *testing.T) {
	gOption = newDefaultOption()
	initCache(2, true)
	nl := []*namespace{
		{
			NamespaceName: "immutable",
			releaseKey:    "immutable",
		},
		{
			NamespaceName: "mutable",
			releaseKey:    "mutable",
		},
	}

	configSets := []map[string]string{
		{
			"imkey1": "imvalue1",
			"imkey2": "imvalue2",
		},
		{
			"key1": "value1",
			"key2": "value2",
		},
	}

	/*firstExpectChanges := []*ChangeEvent{
		{
			Namespace: "immutable",
			Changes: []*ConfigChange{
				{
					Key:        "imkey1",
					NewValue:   "imvalue1",
					ChangeType: 0,
				},
				{
					Key:        "imkey2",
					NewValue:   "imvalue2",
					ChangeType: 0,
				},
			},
		},
		{
			Namespace: "mutable",
			Changes: []*ConfigChange{
				{
					Key:        "key1",
					NewValue:   "value1",
					ChangeType: 0,
				},
				{
					Key:        "key2",
					NewValue:   "value2",
					ChangeType: 0,
				},
			},
		},
	}*/

	configChangeSets := []map[string]string{
		{
			"imkey1": "im modify1",
			"imkey3": "im add3",
		},
		{
			"key1": "modify1",
			"key3": "add3",
		},
	}
	secondExpectChanges := []*ChangeEvent{
		{
			Namespace: "immutable",
			Changes: []*ConfigChange{
				{
					Key:        "imkey3",
					NewValue:   "im add3",
					ChangeType: 0,
				},
				{
					Key:        "imkey1",
					OldValue:   "imvalue1",
					NewValue:   "im modify1",
					ChangeType: 1,
				},
				{
					Key:        "imkey2",
					OldValue:   "imvalue2",
					ChangeType: 2,
				},
			},
		},
		{
			Namespace: "mutable",
			Changes: []*ConfigChange{
				{
					Key:        "key3",
					NewValue:   "add3",
					ChangeType: 0,
				},
				{
					Key:        "key1",
					OldValue:   "value1",
					NewValue:   "modify1",
					ChangeType: 1,
				},
				{
					Key:        "key2",
					OldValue:   "value2",
					ChangeType: 2,
				},
			},
		},
	}

	for i, ns := range nl {
		cfg := &apollo.Config{
			ConnConfig:     apollo.ConnConfig{NamespaceName: ns.NamespaceName},
			Configurations: configSets[i],
		}
		changeEvent := getChangeEvent(cfg)
		updateCache(cfg, ns, changeEvent)

		//assert.Equal(t, firstExpectChanges[i], changeEvent)
		fmt.Printf("nameSpace: %+v\n", changeEvent.Namespace)
		for i := range changeEvent.Changes {
			fmt.Printf("%+v\n", changeEvent.Changes[i])
		}
	}
	b, _ := gConfigCache.Get([]byte("mutable.key1"))
	fmt.Printf("%v\n", string(b))

	for i, ns := range nl {
		cfg := &apollo.Config{
			ConnConfig:     apollo.ConnConfig{NamespaceName: ns.NamespaceName},
			Configurations: configChangeSets[i],
		}
		cl := getChangeEvent(cfg)
		updateCache(cfg, ns, cl)

		assert.Equal(t, secondExpectChanges[i], sortChanges(cl))
		fmt.Printf("nameSpace: %+v\n", cl.Namespace)
		for i := range cl.Changes {
			fmt.Printf("%+v\n", cl.Changes[i])
		}
	}
}
