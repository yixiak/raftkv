package skiplist

import (
	"math/rand/v2"
	"sync"
)

const (
	MAX_LEVEL = 8
)

type ListNode struct {
	Key  int
	Val  string
	Prev *ListNode
	Next []*ListNode

	level int
}

type SkipList struct {
	// 全局表锁
	mu sync.Mutex

	Head *ListNode

	size          int
	level         int
	element_count int

	persist_file string
}

func NewSkipList(file string) *SkipList {
	return &SkipList{
		Head: &ListNode{
			Key:   0,
			Val:   "",
			Prev:  nil,
			Next:  make([]*ListNode, MAX_LEVEL+1),
			level: MAX_LEVEL,
		},
		mu:            sync.Mutex{},
		level:         0,
		element_count: 0,
		persist_file:  file,
	}
}

func NewNode(k int, v string, level int) *ListNode {
	return &ListNode{
		Key:   k,
		Val:   v,
		Prev:  nil,
		Next:  make([]*ListNode, level+1),
		level: level,
	}
}

func (list *SkipList) Insert(k int, v string) int {

	list.mu.Lock()
	defer list.mu.Unlock()

	updates := make([]*ListNode, MAX_LEVEL+1)
	current := list.Head

	for i := list.level; i >= 0; i-- {
		for current.Next[i] != nil && current.Next[i].Key < k {
			current = current.Next[i]
		}
		updates[i] = current
	}

	current = current.Next[0]

	if current != nil && current.Key == k {
		return 1
	}

	if current == nil || current.Key != k {
		randlevel := list.randomLevel()

		if randlevel > list.level {
			for i := list.level + 1; i <= randlevel; i++ {
				updates[i] = list.Head
			}
			list.level = randlevel
		}

		newNode := NewNode(k, v, randlevel)

		for i := 0; i <= randlevel; i++ {
			newNode.Next[i] = updates[i].Next[i]
			updates[i].Next[i] = newNode
		}
		list.element_count++
	}
	return 0
}

func (list *SkipList) Seach(key int) *ListNode {
	current := list.Head

	for i := list.level; i >= 0; i-- {
		for current.Next[i] != nil && current.Next[i].Key < key {
			current = current.Next[i]
		}
	}
	current = current.Next[0]
	if current != nil && current.Key == key {
		return current
	}
	return nil
}

func (list *SkipList) Delete(key int) bool {
	list.mu.Lock()
	defer list.mu.Unlock()

	del := list.Seach(key)
	if del == nil {
		return false
	}

	update := make([]*ListNode, MAX_LEVEL+1)
	current := list.Head
	for i := list.level; i >= 0; i-- {
		for current.Next[i] != nil && current.Next[i].Key < key {
			current = current.Next[i]
		}
		update[i] = current
	}

	for i := 0; i <= list.level; i++ {
		if update[i].Next[i] != del {
			break
		}
		update[i].Next[i] = del.Next[i]
	}

	for list.level > 0 && list.Head.Next[list.level] == nil {
		list.level--
	}
	list.element_count--
	return true
}

func (list *SkipList) randomLevel() int {
	level := 0
	for level < MAX_LEVEL && rand.Float32() < 0.5 {
		level++
	}
	return level
}
