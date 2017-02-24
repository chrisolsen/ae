package tags

import (
	"strings"

	"github.com/chrisolsen/ae/model"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// TagType defines how the tag is created
type TagType int

const tableName = "tags"

// Tags are either custom keys created by the user or auto created
const (
	Custom TagType = iota
	Auto
)

// Tag allows better searching capabilities with AppEngine's Datastore
type Tag struct {
	model.Base
	Value string
	Type  TagType
}

// Generate auto-creates tags for the val passed in with values longer the the minSize
// ex. (apple, 3) => app, appl, apple
// ex. (foo bario, 3) => foo, bar, bari, bario
func Generate(val string, minSize int) []string {
	tags := make(map[string]bool)
	words := strings.Split(val, " ")
	for _, word := range words {
		lc := len(word)
		if lc <= minSize {
			if _, exists := tags[word]; !exists {
				tags[strings.ToLower(word)] = true
			}
			continue
		}
		for i := minSize; i <= lc; i++ {
			if _, exists := tags[word]; !exists {
				tags[strings.ToLower(word[:i])] = true
			}
		}
	}
	tarr := make([]string, len(tags))
	i := 0
	for tag := range tags {
		tarr[i] = tag
		i++
	}
	return tarr
}

// Save saves the tag items with the parent relation
func Save(c context.Context, rawTags []string, tagType TagType, tagTableName string, parentKey *datastore.Key) (int, int, error) {
	var delCount int
	var addCount int

	// get existing
	var existingTags []*Tag
	oldKeys, err := datastore.NewQuery(tagTableName).
		Ancestor(parentKey).
		Filter("Type =", tagType).
		Order("Value").
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
		newTagMap[tag] = true
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
		tg := &Tag{Value: strings.ToLower(t), Type: tagType}
		ky := datastore.NewIncompleteKey(c, tagTableName, parentKey)
		_, err = datastore.Put(c, ky, tg)
		if err != nil {
			return 0, 0, err
		}
	}
	addCount = len(newTags)
	return delCount, addCount, nil
}

// FindKeysByTag returns a list of all the tag's parent datastore keys
func FindKeysByTag(c context.Context, tag, tagTableName string) ([]*datastore.Key, error) {
	filter := strings.ToLower(tag)
	keys, err := datastore.NewQuery(tagTableName).Filter("Value =", filter).KeysOnly().GetAll(c, nil)
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
		if _, ok := existingTags[strings.ToLower(t)]; !ok {
			list = append(list, t)
		}
	}
	return list
}
