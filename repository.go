package agollo

import (
	"github.com/Shonminh/apollo-client/internal/apollo"
	"github.com/Shonminh/apollo-client/internal/logger"
	"github.com/coocood/freecache"
	"github.com/pkg/errors"
	"strings"
)

// exported constants
const (
	EMPTY = ""
	SEP   = "."
)

var (
	gConfigCache *freecache.Cache
)

func initCache(sz int) *freecache.Cache {
	gConfigCache = freecache.NewCache(sz)
	return gConfigCache
}


func GetApolloConfigCache() *freecache.Cache  {
	return gConfigCache
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
			gConfigCache.Set([]byte(ck), []byte(c.NewValue), 0)
		} else if c.ChangeType == DELETED {
			gConfigCache.Del([]byte(ck))
		} else {
			err := errors.New("Wrong ChangeType")
			logger.LogError("Wrong ChangeType %v", c.ChangeType)
			return err
		}
	}
	return nil
}

func getConfigChangeEvent(namespaceName string, configurations map[string]string) []*ConfigChange {
	if (configurations == nil || len(configurations) == 0) && gConfigCache.EntryCount() == 0 {
		return nil
	}

	// get old keys
	nnd := namespaceName + SEP
	mp := map[string]string{}
	it := gConfigCache.NewIterator()
	for en := it.Next(); en != nil; en = it.Next() {
		ck := string(en.Key)
		if strings.HasPrefix(ck, nnd) {
			mp[ck] = string(en.Value)
		}
	}

	changes := make([]*ConfigChange, 0)

	if configurations != nil {
		for k, v := range configurations {
			ck := getCacheKey(namespaceName, k)
			old, ok := mp[ck]
			if ok {
				if old != v {
					changes = append(changes, newModifyConfigChange(k, old, v))
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
		changes = append(changes, newDeletedConfigChange(k, v))
	}

	return changes
}

func getCacheKey(namespaceName string, key string) string {
	return namespaceName + SEP + key
}

func getChangeEvent(ac *apollo.Config) *ChangeEvent {
	// 由于目前只有一个写进程，这个函数调用是没问题的
	cl := getConfigChangeEvent(ac.NamespaceName, ac.Configurations)

	event := &ChangeEvent{
		Namespace: ac.NamespaceName,
		Changes:   cl,
	}
	return event
}
