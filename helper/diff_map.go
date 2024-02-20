package helper

import "net/url"

func DiffMaps(oldMap, newMap map[string]string) map[string]map[string]string {
	diff := map[string]map[string]string{}
	for key, newValue := range newMap {
		newValUnescaped, _ := url.QueryUnescape(newValue)

		oldValue, ok := oldMap[key]
		if !ok {
			diff[key] = map[string]string{
				"old": "",
				"new": newValUnescaped,
			}
		}

		oldValUnescaped, _ := url.QueryUnescape(oldMap[key])
		if newValUnescaped != oldValUnescaped {
			diff[key] = map[string]string{
				"old": oldValue,
				"new": newValue,
			}
		}

		for key, oldValue := range oldMap {
			if _, ok := newMap[key]; !ok {
				diff[key] = map[string]string{
					"old": oldValue,
					"new": "",
				}
			}
		}
	}

	return diff
}
