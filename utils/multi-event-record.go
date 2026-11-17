package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/wisemonkeys-co/event-mapping/types"
)

// splitPathSeparator separates the nested arrays addressed by RecordMap.SplitEventField
// (e.g. "contracts.$.data.products").
const splitPathSeparator = ".$."

// Errors returned by SplitMultiEventRecord. They are always wrapped with context,
// so use errors.Is to check them.
var (
	ErrInvalidSplitEventField = errors.New("invalid split event field")
	ErrSplitArrayNotFound     = errors.New("array to split not found")
	ErrSplitPathNotArray      = errors.New("split path is not an array")
	ErrSplitItemNotObject     = errors.New("split item is not an object")
	ErrSplitEventCopy         = errors.New("could not copy the event")
	ErrSplitSetValue          = errors.New("could not set value in the event")
)

// splitItem addresses one event to be extracted: the item index on each array level
// (root first) or the error that prevented the extraction.
type splitItem struct {
	indexes []int
	err     error
}

func (si splitItem) sufix() string {
	parts := make([]string, len(si.indexes))
	for i, index := range si.indexes {
		parts[i] = strconv.Itoa(index)
	}
	return strings.Join(parts, "_")
}

// SplitMultiEventRecord extracts the events of a record with multiple events, according to
// recordMap.SplitEventField (nested arrays separated by ".$.").
//
// Each extracted event is a copy of the record where every array of the split path is replaced
// by the selected item. When no id field inside the split path has a value, the item sufix
// (e.g. "_0", "_1_2") is appended to the last unique id field (or to the last id field), so each
// extracted event has its own id.
//
// Both returned slices always have the same length: for each index i, either events[i] holds the
// extracted event and errs[i] is nil, or events[i] is nil and errs[i] holds the error. Failures:
//   - before the split (invalid split field, root array not found / not an array): one entry
//     with the error;
//   - on a nested level (item not an object, nested array not found / not an array): one entry
//     with the error for the failed parent item. The other items keep their sufixes.
//
// Empty arrays produce no entries. The received record is never modified, the returned events
// share no references with it nor among themselves, and nothing is logged or published.
func SplitMultiEventRecord(event map[string]any, recordMap types.RecordMap) (events []map[string]any, errs []error) {
	splitEventField := recordMap.SplitEventField
	pathList := strings.Split(splitEventField, splitPathSeparator)
	for _, path := range pathList {
		if path == "" {
			return []map[string]any{nil}, []error{fmt.Errorf("%w: %q", ErrInvalidSplitEventField, splitEventField)}
		}
	}
	// The JSON copy normalizes the record types (e.g. []map[string]any -> []any, numbers -> float64),
	// the same types the extracted events will have.
	normalizedEvent, err := deepCopyMap(event)
	if err != nil {
		return []map[string]any{nil}, []error{err}
	}
	rootList, err := getSplitArray(pathList[0], normalizedEvent)
	if err != nil {
		return []map[string]any{nil}, []error{fmt.Errorf("%w (split path %s)", err, splitEventField)}
	}
	items := collectSplitItems(splitEventField, pathList[1:], rootList, nil)
	events = make([]map[string]any, 0, len(items))
	errs = make([]error, 0, len(items))
	for _, item := range items {
		extractedEvent, err := extractSplitItem(normalizedEvent, pathList, item, recordMap)
		if err != nil {
			events = append(events, nil)
			errs = append(errs, fmt.Errorf("item %s: %w", item.sufix(), err))
			continue
		}
		events = append(events, extractedEvent)
		errs = append(errs, nil)
	}
	return
}

// collectSplitItems walks the nested arrays (read only), returning one item per event to be
// extracted or one item with the error for each parent item that could not be walked.
func collectSplitItems(splitEventField string, pathList []string, list []any, parentIndexes []int) (items []splitItem) {
	for index, arrayItem := range list {
		indexes := make([]int, len(parentIndexes), len(parentIndexes)+1)
		copy(indexes, parentIndexes)
		indexes = append(indexes, index)
		if len(pathList) == 0 {
			items = append(items, splitItem{indexes: indexes})
			continue
		}
		itemMap, ok := arrayItem.(map[string]any)
		if !ok {
			items = append(items, splitItem{
				indexes: indexes,
				err:     fmt.Errorf("%w (found %T, split path %s)", ErrSplitItemNotObject, arrayItem, splitEventField),
			})
			continue
		}
		nestedList, err := getSplitArray(pathList[0], itemMap)
		if err != nil {
			items = append(items, splitItem{
				indexes: indexes,
				err:     fmt.Errorf("%w (split path %s)", err, splitEventField),
			})
			continue
		}
		items = append(items, collectSplitItems(splitEventField, pathList[1:], nestedList, indexes)...)
	}
	return
}

// extractSplitItem copies the record, replaces each array of the split path by the selected item
// and appends the id sufix when needed.
func extractSplitItem(normalizedEvent map[string]any, pathList []string, item splitItem, recordMap types.RecordMap) (map[string]any, error) {
	if item.err != nil {
		return nil, item.err
	}
	eventCopy, err := deepCopyMap(normalizedEvent)
	if err != nil {
		return nil, err
	}
	container := eventCopy
	for level, path := range pathList {
		list, err := getSplitArray(path, container)
		if err != nil {
			return nil, err
		}
		index := item.indexes[level]
		if index >= len(list) {
			return nil, fmt.Errorf("%w: index %d out of range at %s", ErrSplitSetValue, index, path)
		}
		selected := list[index]
		if err = setValueInMap(path, container, selected); err != nil {
			return nil, err
		}
		if level < len(pathList)-1 {
			selectedMap, ok := selected.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%w (found %T at %s)", ErrSplitItemNotObject, selected, path)
			}
			container = selectedMap
		}
	}
	if shouldAppendIdSufix(recordMap.SplitEventField, recordMap.Id, eventCopy) {
		if err = appendIdSufix(recordMap.Id, eventCopy, item.sufix()); err != nil {
			return nil, err
		}
	}
	return eventCopy, nil
}

func getSplitArray(path string, object map[string]any) ([]any, error) {
	value, found := GetValueFromMap(path, object)
	if !found {
		return nil, fmt.Errorf("%w: %s", ErrSplitArrayNotFound, path)
	}
	list, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: %s (found %T)", ErrSplitPathNotArray, path, value)
	}
	return list, nil
}

// shouldAppendIdSufix returns false when an id field inside the split path already has a value
// (the extracted events already have their own ids).
func shouldAppendIdSufix(splitEventField string, idFieldList []types.RecordField, event map[string]any) bool {
	splitPath := strings.ReplaceAll(splitEventField, ".$", "")
	for _, idField := range idFieldList {
		if isInsidePath(idField.RemoteName, splitPath) {
			if value, _ := GetValueFromMap(idField.RemoteName, event); value != nil {
				return false
			}
		}
	}
	return true
}

// isInsidePath compares path segments: "contracts.id" is inside "contracts",
// "contractsBatchId" is not.
func isInsidePath(fieldPath, basePath string) bool {
	return fieldPath == basePath || strings.HasPrefix(fieldPath, basePath+".")
}

// appendIdSufix appends the sufix to the last unique id field (or to the last id field).
func appendIdSufix(idFieldList []types.RecordField, event map[string]any, sufix string) error {
	if len(idFieldList) == 0 {
		return nil
	}
	selectedIndex := len(idFieldList) - 1
	for i := len(idFieldList) - 1; i >= 0; i-- {
		if idFieldList[i].Unique {
			selectedIndex = i
			break
		}
	}
	idPath := idFieldList[selectedIndex].RemoteName
	idPrefix := "<nil>"
	if value, _ := GetValueFromMap(idPath, event); value != nil {
		idPrefix = GetString(value)
	}
	return setValueInMap(idPath, event, fmt.Sprintf("%s_%s", idPrefix, sufix))
}

// setValueInMap sets the value at the dot separated path. Every intermediate node must exist and be an object.
func setValueInMap(path string, object map[string]any, value any) error {
	keys := strings.Split(path, ".")
	container := object
	for _, key := range keys[:len(keys)-1] {
		next, ok := container[key].(map[string]any)
		if !ok {
			return fmt.Errorf("%w: %s (%s is not an object)", ErrSplitSetValue, path, key)
		}
		container = next
	}
	container[keys[len(keys)-1]] = value
	return nil
}

func deepCopyMap(object map[string]any) (map[string]any, error) {
	jsonBytes, err := json.Marshal(object)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrSplitEventCopy, err.Error())
	}
	var objectCopy map[string]any
	if err = json.Unmarshal(jsonBytes, &objectCopy); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrSplitEventCopy, err.Error())
	}
	return objectCopy, nil
}
