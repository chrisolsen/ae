package tags

import (
	"reflect"
	"sort"
	"testing"

	"github.com/chrisolsen/ae/testutils"
	"google.golang.org/appengine/datastore"
)

func TestGenerate(t *testing.T) {
	type test struct {
		val      string
		minSize  int
		expected []string
	}

	tests := []test{
		test{val: "foo", minSize: 3, expected: []string{"foo"}},
		test{val: "foob", minSize: 3, expected: []string{"foo", "foob"}},
		test{val: "foob", minSize: 3, expected: []string{"foo", "foob"}},
		test{val: "fooba", minSize: 3, expected: []string{"foo", "foob", "fooba"}},
		test{val: "foo bar", minSize: 3, expected: []string{"bar", "foo"}},
		test{val: "foo bario", minSize: 3, expected: []string{"bar", "bari", "bario", "foo"}},
		test{val: "foo bario fooz", minSize: 3, expected: []string{"bar", "bari", "bario", "foo", "fooz"}},
	}

	for _, test := range tests {
		tags := Generate(test.val, test.minSize)
		sort.Strings(tags)
		vals := []string{}
		for _, tag := range tags {
			vals = append(vals, tag)
		}
		if !reflect.DeepEqual(vals, test.expected) {
			t.Errorf("%v does not match % v", vals, test.expected)
		}
	}
}

func TestSave(t *testing.T) {
	utils := testutils.T{}
	c := utils.GetContext()
	defer utils.Close()

	// need a parent
	key := datastore.NewIncompleteKey(c, "tags", nil)
	parentKey, _ := datastore.Put(c, key, &Tag{Value: "parent", Type: Auto})

	// init existing tags
	for _, tag := range []string{"rm1", "rm2", "keep"} {
		key := datastore.NewIncompleteKey(c, "tags", parentKey)
		datastore.Put(c, key, &Tag{Value: tag, Type: Custom})
	}

	// save => create new tags
	delCount, addCount, _ := Save(c, []string{"keep", "new"}, Custom, "tags", parentKey)
	if delCount != 2 {
		t.Errorf("%d deleted, expected %d", delCount, 2)
	}
	if addCount != 1 {
		t.Errorf("%d added, expected %d", addCount, 1)
	}
}
