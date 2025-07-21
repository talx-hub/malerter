package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue_Push(t *testing.T) {
	q := New[int]()

	q.Push(1)
	q.Push(2)
	q.Push(3)

	var nums []int
	for q.Len() > 0 {
		nums = append(nums, q.Pop())
	}

	assert.Equal(t, []int{1, 2, 3}, nums)
}
