package cache

import (
	"fmt"
	"sync"
	"time"
)

type node struct {
	key, value interface{}
	prev, next *node
	expire     *time.Time
}

// Cache .
type Cache struct {
	l          sync.RWMutex
	m          map[interface{}]*node
	head, tail *node
	length     int
	maxLength  int
}

// NewCache .
func NewCache(capacity int) *Cache {
	capacity = max(capacity, 1)
	cacheHead := newNode("head", "head")
	cacheTail := newNode("tail", "tail")
	cacheHead.next = cacheTail
	cacheTail.prev = cacheHead
	return &Cache{
		m:         make(map[interface{}]*node),
		head:      cacheHead,
		tail:      cacheTail,
		length:    0,
		maxLength: capacity,
	}
}

// Option .
type Option func(n *node)

// Put .
func (c *Cache) Put(key, value interface{}, opts ...Option) {
	c.l.Lock()
	defer c.l.Unlock()
	if _, ok := c.m[key]; ok {
		c.m[key].prev.next = c.m[key].next
		c.m[key].next.prev = c.m[key].prev
		c.m[key].prev = c.head
		c.m[key].next = c.head.next
		c.head.next = c.m[key]
	} else {
		if c.length == c.maxLength {
			c.remove(c.tail.prev.key)
		} else {
			c.length++
		}
		curHead := newNode(key, value)
		curHead.next = c.head.next
		c.head.next.prev = curHead
		curHead.prev = c.head
		c.head.next = curHead
		c.m[key] = curHead
	}
	for _, o := range opts {
		o(c.m[key])
	}
}

// Get .
func (c *Cache) Get(key interface{}) (interface{}, bool) {
	c.l.RLock()
	defer c.l.RUnlock()
	if _, ok := c.m[key]; !ok {
		return nil, false
	}
	go c.checkExpireNode()
	if c.m[key].expire != nil && c.m[key].expire.Before(time.Now()) {
		return nil, false
	}
	return c.m[key].value, true
}

// List .
func (c *Cache) List() {
	c.l.RLock()
	defer c.l.RUnlock()
	cur := c.head.next
	for cur != c.tail {
		fmt.Printf("%v %v | ", cur.key, cur.value)
		cur = cur.next
	}
	fmt.Println()
	cur = c.tail.prev
	for cur != c.head {
		fmt.Printf("%v %v | ", cur.key, cur.value)
		cur = cur.prev
	}
	fmt.Println()
}

// WithExpire .
func WithExpire(dur time.Duration) Option {
	if dur < 0 {
		dur = time.Duration(0)
	}
	return func(n *node) {
		now := time.Now().Add(dur * 1000000000)
		n.expire = &now
	}
}

func newNode(k, v interface{}) *node {
	return &node{
		key:   k,
		value: v,
	}
}

func (c *Cache) checkExpireNode() {
	c.l.Lock()
	defer c.l.Unlock()
	now := time.Now()
	for _, v := range c.m {
		if v.expire != nil && v.expire.Before(now) {
			c.remove(v.key)
		}
	}
}

func (c *Cache) remove(key interface{}) {
	c.m[key].next.prev = c.m[key].prev
	c.m[key].prev.next = c.m[key].next
	c.m[key].next, c.m[key].prev = nil, nil
	delete(c.m, key)
}
