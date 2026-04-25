package mathtext

import "sync"

// Cache stores parsed MathText expressions and optionally final layouts. Layout
// entries are keyed by a caller-supplied MeasurementKey because final layout
// depends on renderer-specific text metrics and font resolution.
type Cache struct {
	mu      sync.RWMutex
	parsed  map[string]mathLayoutNode
	layouts map[layoutCacheKey]MathTextLayout
}

type layoutCacheKey struct {
	kind           string
	text           string
	size           float64
	fontKey        string
	measurementKey string
}

var defaultCache = NewCache()

// NewCache creates an empty MathText cache.
func NewCache() *Cache {
	return &Cache{
		parsed:  map[string]mathLayoutNode{},
		layouts: map[layoutCacheKey]MathTextLayout{},
	}
}

// DefaultCache returns the process-wide cache used for renderer-independent
// parsing. Callers that enable layout caching should normally own their own
// cache or provide a MeasurementKey that isolates renderer metric behavior.
func DefaultCache() *Cache {
	return defaultCache
}

// Clear removes all cached parsed expressions and layouts.
func (c *Cache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.parsed = map[string]mathLayoutNode{}
	c.layouts = map[layoutCacheKey]MathTextLayout{}
}

// Stats returns the current parsed-expression and layout entry counts.
func (c *Cache) Stats() (parsed, layouts int) {
	if c == nil {
		return 0, 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.parsed), len(c.layouts)
}

func (c *Cache) parsedNode(expr string) (mathLayoutNode, bool) {
	if c == nil {
		return mathLayoutNode{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	node, ok := c.parsed[expr]
	return node, ok
}

func (c *Cache) storeParsedNode(expr string, node mathLayoutNode) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.parsed[expr] = node
}

func (c *Cache) layout(key layoutCacheKey) (MathTextLayout, bool) {
	if c == nil {
		return MathTextLayout{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	layout, ok := c.layouts[key]
	if !ok {
		return MathTextLayout{}, false
	}
	return cloneLayout(layout), true
}

func (c *Cache) storeLayout(key layoutCacheKey, layout MathTextLayout) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.layouts[key] = cloneLayout(layout)
}

func cloneLayout(layout MathTextLayout) MathTextLayout {
	layout.Runs = append([]MathTextLayoutRun(nil), layout.Runs...)
	layout.Rules = append([]MathTextLayoutRule(nil), layout.Rules...)
	return layout
}
