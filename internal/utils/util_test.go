package utils

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortIdList(t *testing.T) {
	idList := IdList{12, 4, 8, 23, 9}
	sorted := IdList{4, 8, 9, 12, 23}
	sort.Sort(idList)
	for i, _ := range idList {
		assert.Equal(t, sorted[i], idList[i], fmt.Sprintf("Test id at index %d", i))
	}
}
