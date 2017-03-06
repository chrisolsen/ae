package tags

import (
	"strings"

	"github.com/chrisolsen/ae/model"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const tableName = "tags"

// Tag allows better searching capabilities with AppEngine's Datastore
type Tag struct {
	model.Base
	Value  string
	Type   string
	Public bool
}

// Save saves the tag items with the parent relation
func Save(c context.Context, rawTags []string, tagType string, public bool, parentKey *datastore.Key) (int, int, error) {
	var delCount int
	var addCount int

	// get existing
	var existingTags []*Tag
	oldKeys, err := datastore.NewQuery(tableName).
		Ancestor(parentKey).
		Filter("Type =", tagType).
		Filter("Public =", public).
		GetAll(c, &existingTags)
	if err != nil {
		return 0, 0, err
	}
	for i, key := range oldKeys {
		existingTags[i].Key = key
	}

	// delete old tags
	newTagMap := make(map[string]bool)
	for _, tag := range rawTags {
		newTagMap[strings.ToLower(tag)] = true
	}

	rmTagKeys := getTagsToRemove(existingTags, newTagMap)
	err = datastore.DeleteMulti(c, rmTagKeys)
	if err != nil {
		return 0, 0, err
	}
	delCount = len(rmTagKeys)

	// insert new tags
	existingTagMap := make(map[string]bool)
	for _, tag := range existingTags {
		existingTagMap[tag.Value] = true
	}
	newTags := getTagsToAdd(rawTags, existingTagMap)
	for _, t := range newTags {
		tg := &Tag{Value: strings.ToLower(t), Type: tagType, Public: public}
		ky := datastore.NewIncompleteKey(c, tableName, parentKey)
		_, err = datastore.Put(c, ky, tg)
		if err != nil {
			return 0, 0, err
		}
	}
	addCount = len(newTags)
	return delCount, addCount, nil
}

// FindKeysByTag returns a list of all the tag's parent datastore keys
func FindKeysByTag(c context.Context, tag, tagType string, parentKey *datastore.Key, offset, limit int) ([]*datastore.Key, error) {
	filter := strings.ToLower(tag)
	q := datastore.NewQuery(tableName)
	if parentKey != nil {
		q = q.Ancestor(parentKey)
	}
	keys, err := q.
		Filter("Type =", tagType).
		Filter("Value =", filter).
		Offset(offset).
		Limit(limit).
		KeysOnly().
		GetAll(c, nil)

	if err != nil {
		return nil, err
	}
	parentKeys := make([]*datastore.Key, len(keys))
	for i, k := range keys {
		parentKeys[i] = k.Parent()
	}
	return parentKeys, err
}

func getTagsToRemove(currentTags []*Tag, newTags map[string]bool) []*datastore.Key {
	var list []*datastore.Key
	for _, t := range currentTags {
		if _, ok := newTags[strings.ToLower(t.Value)]; !ok {
			list = append(list, t.Key)
		}
	}
	return list
}

func getTagsToAdd(rawTags []string, existingTags map[string]bool) []string {
	var list []string
	for _, t := range rawTags {
		lt := strings.ToLower(t)
		if _, ok := existingTags[lt]; !ok {
			list = append(list, lt)
		}
	}
	return list
}
