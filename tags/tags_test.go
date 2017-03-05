package tags

import (
	"testing"

	"github.com/chrisolsen/ae/testutils"
	"google.golang.org/appengine/datastore"
)

func TestSave(t *testing.T) {
	utils := testutils.T{}
	c := utils.GetContext()
	defer utils.Close()

	// need a parent
	key := datastore.NewIncompleteKey(c, "tags", nil)
	parentKey, _ := datastore.Put(c, key, &Tag{Value: "parent", Type: "parent"})

	// init existing tags
	for _, tag := range []string{"rm1", "rm2", "keep"} {
		key := datastore.NewIncompleteKey(c, "tags", parentKey)
		datastore.Put(c, key, &Tag{Value: tag, Type: "sometype", Public: true})
	}

	// save => create new tags
	delCount, addCount, _ := Save(c, []string{"keep", "new"}, "sometype", true, parentKey)
	if delCount != 2 {
		t.Errorf("%d deleted, expected %d", delCount, 2)
	}
	if addCount != 1 {
		t.Errorf("%d added, expected %d", addCount, 1)
	}
}
