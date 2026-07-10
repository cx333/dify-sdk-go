package client

import (
	"testing"
)

// ---- SplitThink ----

func TestSplitThink_Basic(t *testing.T) {
	thinking, clean := SplitThink("<think>Let me think...</think>Done!")
	if thinking != "Let me think..." {
		t.Errorf("thinking = %q", thinking)
	}
	if clean != "Done!" {
		t.Errorf("clean = %q", clean)
	}
}

func TestSplitThink_NoTag(t *testing.T) {
	thinking, clean := SplitThink("Just an answer, no thinking")
	if thinking != "" {
		t.Errorf("thinking = %q, want empty", thinking)
	}
	if clean != "Just an answer, no thinking" {
		t.Errorf("clean = %q", clean)
	}
}

func TestSplitThink_OnlyOpenTag(t *testing.T) {
	thinking, clean := SplitThink("Text <think>unclosed")
	if thinking != "" {
		t.Errorf("thinking = %q, want empty", thinking)
	}
	if clean != "Text <think>unclosed" {
		t.Errorf("clean = %q", clean)
	}
}

func TestSplitThink_EmptyThinkBlock(t *testing.T) {
	thinking, clean := SplitThink("before<think></think>after")
	if thinking != "" {
		t.Errorf("thinking = %q, want empty", thinking)
	}
	// empty think block removes itself with no extra space
	if clean != "beforeafter" {
		t.Errorf("clean = %q", clean)
	}
}

func TestSplitThink_TextBeforeAndAfter(t *testing.T) {
	thinking, clean := SplitThink("Intro. <think>reasoning here</think> Conclusion.")
	if thinking != "reasoning here" {
		t.Errorf("thinking = %q", thinking)
	}
	if clean != "Intro.  Conclusion." {
		t.Errorf("clean = %q", clean)
	}
}

func TestSplitThink_TrimsWhitespace(t *testing.T) {
	thinking, clean := SplitThink("<think>  \n reasoning \t </think>  answer  ")
	if thinking != "reasoning" {
		t.Errorf("thinking = %q", thinking)
	}
	if clean != "answer" {
		t.Errorf("clean = %q", clean)
	}
}

// ---- ThinkParser.Feed ----

func TestThinkParser_CompleteInOneChunk(t *testing.T) {
	var p ThinkParser
	events := p.Feed("Hello <think>processing</think> World")

	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	assertEvent(t, events[0], "answer", "Hello ")
	assertEvent(t, events[1], "thinking", "processing")
	assertEvent(t, events[2], "answer", " World")
}

func TestThinkParser_SplitAcrossChunks(t *testing.T) {
	var p ThinkParser

	events := p.Feed("Start <thin") // tag boundary
	if len(events) != 1 {
		t.Fatalf("chunk 1: got %d events, want 1", len(events))
	}
	assertEvent(t, events[0], "answer", "Start ")

	// chunk contains think content up to partial closing tag "</thi"
	events = p.Feed("k>reasoning goes here</thi")
	if len(events) != 1 {
		t.Fatalf("chunk 2: got %d events, want 1 (safe emitted before partial </think>)", len(events))
	}
	assertEvent(t, events[0], "thinking", "reasoning goes here")

	events = p.Feed("nk> end")
	if len(events) != 1 {
		t.Fatalf("chunk 3: got %d events, want 1", len(events))
	}
	assertEvent(t, events[0], "answer", " end")
}

func TestThinkParser_NoThinkTag(t *testing.T) {
	var p ThinkParser
	events := p.Feed("Plain answer without any tags")
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	assertEvent(t, events[0], "answer", "Plain answer without any tags")
}

func TestThinkParser_EmptyChunk(t *testing.T) {
	var p ThinkParser
	events := p.Feed("")
	if len(events) != 0 {
		t.Fatalf("got %d events, want 0", len(events))
	}
}

func TestThinkParser_MultipleThinkBlocks(t *testing.T) {
	var p ThinkParser
	events := p.Feed("A <think>t1</think> B <think>t2</think> C")

	if len(events) != 5 {
		t.Fatalf("got %d events, want 5", len(events))
	}
	assertEvent(t, events[0], "answer", "A ")
	assertEvent(t, events[1], "thinking", "t1")
	assertEvent(t, events[2], "answer", " B ")
	assertEvent(t, events[3], "thinking", "t2")
	assertEvent(t, events[4], "answer", " C")
}

func TestThinkParser_UnclosedThinkTag(t *testing.T) {
	var p ThinkParser
	events := p.Feed("Before <think>never ends...")

	// "Before " emitted as answer, then "never ends..." buffered in think mode
	// safeTail emits it since "..." doesn't match any "</think>" prefix
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	assertEvent(t, events[0], "answer", "Before ")
	assertEvent(t, events[1], "thinking", "never ends...")

	// Subsequent chunk stays buffered
	events = p.Feed(" more thinking")
	if len(events) != 1 {
		t.Fatalf("subsequent chunk: got %d events, want 1 (still buffered, safeTail emits)", len(events))
	}
	assertEvent(t, events[0], "thinking", " more thinking")
}

func TestThinkParser_OnlyClosingTag(t *testing.T) {
	var p ThinkParser
	// </think> without opening <think> should be treated as answer text
	events := p.Feed("Text </think> more")

	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	assertEvent(t, events[0], "answer", "Text </think> more")
}

func TestThinkParser_ImmediateThinkTag(t *testing.T) {
	var p ThinkParser
	events := p.Feed("<think>thinking only, no answer after")

	// safeTail emits the full content since no </think> found and no partial prefix
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1 (content emitted via safeTail)", len(events))
	}
	assertEvent(t, events[0], "thinking", "thinking only, no answer after")
}

func TestThinkParser_StreamSimulation(t *testing.T) {
	// Simulate realistic streaming where tags and content arrive in word-level chunks.
	// The parser handles partial tags buffered across chunks correctly.
	var p ThinkParser

	// Answer text arrives, then think tag starts
	events := p.Feed("Hello")
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	assertEvent(t, events[0], "answer", "Hello")

	events = p.Feed("<think>reasoning step 1,")
	// safeTail emits since "," doesn't match any "</think>" prefix
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1 (safeTail emitted)", len(events))
	}
	assertEvent(t, events[0], "thinking", "reasoning step 1,")

	events = p.Feed(" step 2</think>")
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1 (thinking content before </think>)", len(events))
	}
	assertEvent(t, events[0], "thinking", " step 2")

	events = p.Feed("Final answer.")
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	assertEvent(t, events[0], "answer", "Final answer.")
}

// ---- ThinkParser.Reset ----

func TestThinkParser_Reset(t *testing.T) {
	var p ThinkParser
	p.Feed("Start <think>processing")

	p.Reset()

	events := p.Feed("Fresh start")
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	assertEvent(t, events[0], "answer", "Fresh start")
}

// ---- safeTail ----

func TestSafeTail_FullString(t *testing.T) {
	// No partial tag prefix at end, should return full length
	if got := safeTail("hello world", "<think>"); got != 11 {
		t.Errorf("safeTail = %d, want 11", got)
	}
}

func TestSafeTail_PartialTagPrefix(t *testing.T) {
	// Ends with "<th" which is a prefix of "<think>"
	if got := safeTail("hello <th", "<think>"); got != 6 {
		t.Errorf("safeTail = %d, want 6", got)
	}
}

func TestSafeTail_SingleCharTagPrefix(t *testing.T) {
	// Ends with "<" which is a prefix of "<think>"
	if got := safeTail("text<", "<think>"); got != 4 {
		t.Errorf("safeTail = %d, want 4", got)
	}
}

func TestSafeTail_EmptyString(t *testing.T) {
	if got := safeTail("", "<think>"); got != 0 {
		t.Errorf("safeTail = %d, want 0", got)
	}
}

func TestSafeTail_TagAtStart(t *testing.T) {
	// Full tag present, no partial truncation
	if got := safeTail("<think>content", "<think>"); got != 14 {
		t.Errorf("safeTail = %d, want 14", got)
	}
}

// ---- Helpers ----

func assertEvent(t *testing.T, ev ThinkEvent, wantType, wantContent string) {
	t.Helper()
	if ev.Type != wantType {
		t.Errorf("event type = %q, want %q", ev.Type, wantType)
	}
	if ev.Content != wantContent {
		t.Errorf("event content = %q, want %q", ev.Content, wantContent)
	}
}
