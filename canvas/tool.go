package canvas

import "sync"

// ToolArgs carries optional execution arguments for a tool action.
type ToolArgs struct {
	Path    string
	Payload any
}

// Tool defines one tool-manager action.
type Tool interface {
	ID() string
	Execute(ToolArgs) error
}

// ToolFunc adapts a function to a Tool.
type ToolFunc struct {
	Name string
	Run  func(ToolArgs) error
}

// ID reports the stable tool identifier.
func (t ToolFunc) ID() string { return t.Name }

// Execute runs the adapted tool function.
func (t ToolFunc) Execute(args ToolArgs) error {
	if t.Run == nil {
		return nil
	}
	return t.Run(args)
}

// ToolManager stores and runs named tools.
type ToolManager struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewToolManager creates an empty tool registry.
func NewToolManager() *ToolManager {
	return &ToolManager{tools: make(map[string]Tool)}
}

// Register adds or replaces a tool.
func (m *ToolManager) Register(tool Tool) {
	if m == nil || tool == nil || tool.ID() == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tools == nil {
		m.tools = make(map[string]Tool)
	}
	m.tools[tool.ID()] = tool
}

// Tool returns the registered tool by id.
func (m *ToolManager) Tool(id string) (Tool, bool) {
	if m == nil {
		return nil, false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	tool, ok := m.tools[id]
	return tool, ok
}

// Execute runs a registered tool by id.
func (m *ToolManager) Execute(id string, args ToolArgs) error {
	tool, ok := m.Tool(id)
	if !ok {
		return nil
	}
	return tool.Execute(args)
}
