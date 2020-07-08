// Forked from https://github.com/zouyx/agollo

package agollo

import (
	"github.com/Shonminh/apollo-client/internal/apollo"
	"github.com/Shonminh/apollo-client/internal/logger"
	"github.com/coocood/freecache"
	json "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"log"
	"strings"
	"sync"
)

// exported constants
const (
	EMPTY = ""
	SEP   = "."
)

var (
	gConfigCache     *cache
	cacheMutex       sync.Mutex
	gIgnoreNameSpace bool
)

func initCache(sz int, ignore bool) *cache {
	gConfigCache = &cache{
		fc: freecache.NewCache(sz),
	}
	gIgnoreNameSpace = ignore
	return gConfigCache
}

func GetConfigCacheMap() map[string]string {
	configMap := make(map[string]string)
	cacheMutex.Lock()
	it := gConfigCache.NewIterator()
	for en := it.Next(); en != nil; en = it.Next() {
		key := string(en.Key)
		value := string(en.Value)
		configMap[key] = value
	}
	cacheMutex.Unlock()
	return configMap
}

func GetConfigByKey(key string) (string, error) {
	cacheMutex.Lock()
	value, err := gConfigCache.Get([]byte(key))
	cacheMutex.Unlock()
	if err != nil {
		return EMPTY, errors.WithMessage(err, "get value")
	}
	return string(value), nil
}

func Cleanup() {
	cacheMutex.Lock()
	gConfigCache.Clear()
	cacheMutex.Unlock()
}

func updateCache(ac *apollo.Config, ns *namespace, event *ChangeEvent) {
	if ac == nil || ns == nil {
		// nothing changed
		return
	}

	ns.releaseKey = ac.ReleaseKey
	doUpdateCache(event)

	// write config file async
	path := getConfigPath(ac.NamespaceName)
	_ = writeConfigFile(ac, path)
}

func pushChange(ns *namespace, event *ChangeEvent) {
	if len(event.Changes) > 0 {
		if gCallback != nil {
			err := gCallback(event)
			if err != nil {
				logger.LogError("Callback config change handler fail: %v", err)
				return
			}
		}
	}
}
func doUpdateCache(event *ChangeEvent) error {
	ns := event.Namespace
	var ck string
	for _, c := range event.Changes {
		ck = getCacheKey(ns, c.Key)
		if c.ChangeType == MODIFIED || c.ChangeType == ADDED {
			cacheMutex.Lock()
			gConfigCache.Set([]byte(ck), &element{Val: c.NewValue, NameSpace: ns}, 0)
			cacheMutex.Unlock()
		} else if c.ChangeType == DELETED {
			cacheMutex.Lock()
			gConfigCache.Del([]byte(ck))
			cacheMutex.Unlock()
		} else {
			err := errors.New("Wrong ChangeType")
			logger.LogError("Wrong ChangeType %v", c.ChangeType)
			return err
		}
	}
	return nil
}

func getConfigChangeEvent(namespaceName string, configurations map[string]string) []*ConfigChange {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if (configurations == nil || len(configurations) == 0) && gConfigCache.EntryCount() == 0 {
		return nil
	}

	// get old keys
	nnd := namespaceName + SEP
	mp := make(map[string]element)
	it := gConfigCache.NewIterator()
	for en := it.Next(); en != nil; en = it.Next() {
		ck := string(en.Key)
		if strings.HasPrefix(ck, nnd) {
			mp[ck] = gConfigCache.Unmarshal(en.Value)
		}
	}

	changes := make([]*ConfigChange, 0)

	if configurations != nil {
		for k, v := range configurations {
			ck := getCacheKey(namespaceName, k)
			old, ok := mp[ck]
			if ok {
				if old.Val != v {
					changes = append(changes, newModifyConfigChange(k, old.Val, v))
				}
				delete(mp, ck)
			} else {
				changes = append(changes, newAddConfigChange(k, v))
			}
		}
	}

	// remove del keys
	for ck, v := range mp {
		k := ck[strings.Index(ck, SEP)+1:]
		changes = append(changes, newDeletedConfigChange(k, v.Val))
	}

	return changes
}

func getChangeEventWithIgnore(namespaceName string, configurations map[string]string) []*ConfigChange {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if (configurations == nil || len(configurations) == 0) && gConfigCache.EntryCount() == 0 {
		return nil
	}

	// get old keys
	mp := make(map[string]element)
	it := gConfigCache.NewIterator()
	for en := it.Next(); en != nil; en = it.Next() {
		ck := string(en.Key)
		mp[ck] = gConfigCache.Unmarshal(en.Value)
	}

	configChanges := make([]*ConfigChange, 0)
	if configurations != nil {
		for k, v := range configurations {
			ck := getCacheKey(namespaceName, k) // immutable key
			old, ok := mp[ck]                   // all key
			if ok {
				if old.Val != v {
					configChanges = append(configChanges, newModifyConfigChange(k, old.Val, v))
				}
				delete(mp, ck) // remain mutable
			} else {
				configChanges = append(configChanges, newAddConfigChange(k, v))
			}
		}
	}

	// remove del keys
	for ck, v := range mp {
		if v.NameSpace != namespaceName {
			delete(mp, ck) // remove config not included this nameSpace, avoid match delete configuration
			continue
		}
		configChanges = append(configChanges, newDeletedConfigChange(ck, v.Val))
	}

	return configChanges
}

func getCacheKey(namespaceName string, key string) string {
	if gIgnoreNameSpace {
		return key
	}
	return namespaceName + SEP + key
}

func getChangeEvent(ac *apollo.Config) *ChangeEvent {
	// Currently, only one goroutine will write memory
	var cl []*ConfigChange
	if gIgnoreNameSpace {
		cl = getChangeEventWithIgnore(ac.NamespaceName, ac.Configurations)
	} else {
		cl = getConfigChangeEvent(ac.NamespaceName, ac.Configurations)
	}

	event := &ChangeEvent{
		Namespace: ac.NamespaceName,
		Changes:   cl,
	}
	return event
}

type element struct {
	Val       string `json:"val"`
	NameSpace string `json:"name_space"`
}

type cache struct {
	fc *freecache.Cache
}

func (c *cache) Get(key []byte) (value []byte, err error) {
	var bytes []byte
	if bytes, err = c.fc.Get(key); err != nil {
		return nil, err
	}
	e := c.Unmarshal(bytes)
	return []byte(e.Val), nil
}

func (c *cache) Set(key []byte, e *element, expireSeconds int) (err error) {
	var bytes []byte
	if bytes, err = c.Marshal(e); err != nil {
		return err
	}
	return c.fc.Set(key, bytes, expireSeconds)
}

func (c *cache) Del(key []byte) (affected bool) {
	return c.fc.Del(key)
}

func (c *cache) NewIterator() *freecache.Iterator {
	return c.fc.NewIterator()
}

func (c *cache) EntryCount() int64 {
	return c.fc.EntryCount()
}

func (c *cache) Clear() {
	c.fc.Clear()
}

func (c *cache) Marshal(e *element) ([]byte, error) {
	return json.Marshal(e)
}

// use element, not *element, to avoid allocation in Get.
func (c *cache) Unmarshal(bytes []byte) (e element) {
	var val element
	if err := json.Unmarshal(bytes, &val); err != nil {
		log.Printf("cache Unmarshal err: %+v", err)
	}
	return val
}
