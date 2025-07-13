package ccfeedback

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
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

// PostToolUseMessage is sent after a tool has been executed
type PostToolUseMessage struct {
	BaseHookMessage
	ToolName   string                 `json:"tool_name"`
	ToolInput  map[string]interface{} `json:"tool_input"`
	ToolOutput interface{}            `json:"tool_output,omitempty"`
	ToolError  string                 `json:"tool_error,omitempty"`
}

// NotificationMessage is sent for system notifications
type NotificationMessage struct {
	BaseHookMessage
	NotificationType string `json:"notification_type"`
	Message          string `json:"message"`
}

// StopMessage is sent when the main agent finishes
type StopMessage struct {
	BaseHookMessage
	Reason       string `json:"reason,omitempty"`
	FinalMessage string `json:"final_message,omitempty"`
}

// SubagentStopMessage is sent when a subagent completes
type SubagentStopMessage struct {
	BaseHookMessage
	SubagentID   string `json:"subagent_id"`
	SubagentName string `json:"subagent_name"`
	Result       string `json:"result,omitempty"`
}

// PreCompactMessage is sent before context compression
type PreCompactMessage struct {
	BaseHookMessage
	CurrentTokens int `json:"current_tokens"`
	TargetTokens  int `json:"target_tokens"`
}

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
