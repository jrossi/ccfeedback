package gismo

import "encoding/json"

// HookMessage is the interface implemented by all hook message types
type HookMessage interface {
	GetBaseMessage() BaseHookMessage
	EventName() HookEventName
}

// HookEventName represents the type of hook event
type HookEventName string

const (
	PreToolUseEvent   HookEventName = "PreToolUse"
	PostToolUseEvent  HookEventName = "PostToolUse"
	NotificationEvent HookEventName = "Notification"
	StopEvent         HookEventName = "Stop"
	SubagentStopEvent HookEventName = "SubagentStop"
	PreCompactEvent   HookEventName = "PreCompact"
)

// BaseHookMessage contains common fields for all hook messages
// Optimized for go-json bitmap optimization (≤16 fields)
type BaseHookMessage struct {
	SessionID      string        `json:"session_id"`
	TranscriptPath string        `json:"transcript_path"`
	HookEventName  HookEventName `json:"hook_event_name"`
}

// PreToolUseMessage is sent before a tool is executed
type PreToolUseMessage struct {
	BaseHookMessage
	ToolName  string                     `json:"tool_name"`
	ToolInput map[string]json.RawMessage `json:"tool_input"`
}

func (m PreToolUseMessage) GetBaseMessage() BaseHookMessage { return m.BaseHookMessage }
func (m PreToolUseMessage) EventName() HookEventName        { return PreToolUseEvent }

// PostToolUseMessage is sent after a tool has been executed
type PostToolUseMessage struct {
	BaseHookMessage
	ToolName   string                     `json:"tool_name"`
	ToolInput  map[string]json.RawMessage `json:"tool_input"`
	ToolOutput json.RawMessage            `json:"tool_output,omitempty"`
	ToolError  string                     `json:"tool_error,omitempty"`
}

func (m PostToolUseMessage) GetBaseMessage() BaseHookMessage { return m.BaseHookMessage }
func (m PostToolUseMessage) EventName() HookEventName        { return PostToolUseEvent }

// NotificationMessage is sent for system notifications
type NotificationMessage struct {
	BaseHookMessage
	NotificationType string `json:"notification_type"`
	Message          string `json:"message"`
}

func (m NotificationMessage) GetBaseMessage() BaseHookMessage { return m.BaseHookMessage }
func (m NotificationMessage) EventName() HookEventName        { return NotificationEvent }

// StopMessage is sent when the main agent finishes
type StopMessage struct {
	BaseHookMessage
	Reason       string `json:"reason,omitempty"`
	FinalMessage string `json:"final_message,omitempty"`
}

func (m StopMessage) GetBaseMessage() BaseHookMessage { return m.BaseHookMessage }
func (m StopMessage) EventName() HookEventName        { return StopEvent }

// SubagentStopMessage is sent when a subagent completes
type SubagentStopMessage struct {
	BaseHookMessage
	SubagentID   string `json:"subagent_id"`
	SubagentName string `json:"subagent_name"`
	Result       string `json:"result,omitempty"`
}

func (m SubagentStopMessage) GetBaseMessage() BaseHookMessage { return m.BaseHookMessage }
func (m SubagentStopMessage) EventName() HookEventName        { return SubagentStopEvent }

// PreCompactMessage is sent before context compression
type PreCompactMessage struct {
	BaseHookMessage
	CurrentTokens int `json:"current_tokens"`
	TargetTokens  int `json:"target_tokens"`
}

func (m PreCompactMessage) GetBaseMessage() BaseHookMessage { return m.BaseHookMessage }
func (m PreCompactMessage) EventName() HookEventName        { return PreCompactEvent }

// HookResponse represents the response from a hook
type HookResponse struct {
	Continue       *bool  `json:"continue,omitempty"`
	StopReason     string `json:"stopReason,omitempty"`
	SuppressOutput *bool  `json:"suppressOutput,omitempty"`
	Decision       string `json:"decision,omitempty"` // For PreToolUse: "block" or "approve"
	Reason         string `json:"reason,omitempty"`   // For PreToolUse: reason for decision
	Message        string `json:"message,omitempty"`  // User-visible message
}

// ExitCode represents the hook exit status
type ExitCode int

const (
	ExitSuccess  ExitCode = 0 // Success - stdout shown in transcript
	ExitBlocking ExitCode = 2 // Blocking error - stderr processed by Claude
)
