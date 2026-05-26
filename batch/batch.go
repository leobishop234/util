package batch

import (
	"github.com/leobishop234/util/srverr"
)

func GetMapValue[K comparable, V any](itemMap map[K]V, key K) (item V, err error) {
	if itemMap == nil {
		return item, srverr.New(srverr.ErrCodeInternal, "unexpected nil map", nil)
	}

	item, ok := itemMap[key]
	if !ok {
		return item, srverr.New(srverr.ErrCodeNotFound, "key not found", nil)
	}

	return item, nil
}

func GetMapSlice[K comparable, V any](itemsMap map[K][]V, key K) ([]V, error) {
	if itemsMap == nil {
		return nil, srverr.New(srverr.ErrCodeInternal, "unexpected nil map", nil)
	}

	items, ok := itemsMap[key]
	if !ok {
		return nil, srverr.New(srverr.ErrCodeNotFound, "key not found", nil)
	}

	return items, nil
}

func TransformMapValues[K comparable, V, O any](itemsMap map[K]V, parseFn func(V) O) map[K]O {
	if itemsMap == nil {
		return nil
	}

	output := make(map[K]O, len(itemsMap))
	for key := range itemsMap {
		output[key] = parseFn(itemsMap[key])
	}
	return output
}

func TransformMapSlices[K comparable, V, O any](itemsMap map[K][]V, parseFn func([]V) []O) map[K][]O {
	if itemsMap == nil {
		return nil
	}

	output := make(map[K][]O, len(itemsMap))
	for key := range itemsMap {
		output[key] = parseFn(itemsMap[key])
	}
	return output
}

func IndexBy[I, K comparable](items []I, keyFn func(item *I) K) (map[K]I, error) {
	if items == nil {
		return nil, nil
	}

	itemsMap := make(map[K]I, len(items))
	for _, item := range items {
		key := keyFn(&item)
		if _, exists := itemsMap[key]; exists {
			return nil, srverr.New(srverr.ErrCodeInternal, "duplicate key", nil)
		}
		itemsMap[key] = item
	}
	return itemsMap, nil
}

func IndexTo[I, K comparable, O any](items []I, keyFn func(item *I) K, parseFn func(item *I) O) (map[K]O, error) {
	if items == nil {
		return nil, nil
	}

	itemsMap := make(map[K]O, len(items))
	for _, item := range items {
		key := keyFn(&item)
		if _, exists := itemsMap[key]; exists {
			return nil, srverr.New(srverr.ErrCodeInternal, "duplicate key", nil)
		}
		itemsMap[key] = parseFn(&item)
	}
	return itemsMap, nil
}

func GroupBy[I, K comparable](items []I, keyFn func(item *I) K) (map[K][]I, error) {
	if items == nil {
		return nil, nil
	}

	itemsMap := make(map[K][]I, len(items))
	for _, item := range items {
		key := keyFn(&item)

		if _, exists := itemsMap[key]; !exists {
			itemsMap[key] = []I{}
		}

		itemsMap[key] = append(itemsMap[key], item)
	}
	return itemsMap, nil
}

func GroupTo[I, K comparable, O any](items []I, keyFn func(item *I) K, parseFn func(item *I) O) (map[K][]O, error) {
	if items == nil {
		return nil, nil
	}

	itemsMap := make(map[K][]O, len(items))
	for _, item := range items {
		key := keyFn(&item)

		if _, exists := itemsMap[key]; !exists {
			itemsMap[key] = []O{}
		}

		itemsMap[key] = append(itemsMap[key], parseFn(&item))
	}
	return itemsMap, nil
}

func TransformSlice[I, O any](items []I, parseFn func(item I) O) []O {
	if items == nil {
		return nil
	}

	out := make([]O, len(items))
	for i := range items {
		out[i] = parseFn(items[i])
	}
	return out
}

func Unique[K comparable](items []K) []K {
	if items == nil {
		return nil
	}

	seen := make(map[K]bool, len(items))
	out := make([]K, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}

	return out
}
