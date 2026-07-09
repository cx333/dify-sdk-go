package client

import "strings"

// SplitThink 从回答中分离 <think>...</think> 思考过程和最终回复。
// 返回 (thinking, cleanAnswer)。
func SplitThink(answer string) (thinking, clean string) {
	const startTag = "<think>"
	const endTag = "</think>"

	start := strings.Index(answer, startTag)
	if start == -1 {
		return "", answer
	}

	end := strings.Index(answer, endTag)
	if end == -1 {
		return "", answer
	}

	thinking = answer[start+len(startTag) : end]
	clean = strings.TrimSpace(answer[:start] + answer[end+len(endTag):])
	return strings.TrimSpace(thinking), clean
}

// ThinkEvent 流式解析后的结构化事件。
type ThinkEvent struct {
	Type    string // "thinking" | "answer"
	Content string // 文本片段
}

// ThinkParser 流式 <think> 标签解析状态机。
// 处理跨 chunk 的标签边界，将原始文本流转换为带类型的事件流。
type ThinkParser struct {
	inThink bool
	buf     string
}

// Feed 喂入原始文本 chunk，返回识别出的结构化事件。
// 未闭合标签前的文本会暂存，待标签闭合后一次性发出。
func (p *ThinkParser) Feed(chunk string) []ThinkEvent {
	p.buf += chunk
	var events []ThinkEvent

	for {
		if !p.inThink {
			idx := strings.Index(p.buf, "<think>")
			if idx == -1 {
				safe := safeTail(p.buf, "<think>")
				if safe > 0 {
					events = append(events, ThinkEvent{Type: "answer", Content: p.buf[:safe]})
					p.buf = p.buf[safe:]
				}
				break
			}
			if idx > 0 {
				events = append(events, ThinkEvent{Type: "answer", Content: p.buf[:idx]})
			}
			p.buf = p.buf[idx+len("<think>"):]
			p.inThink = true
		} else {
			idx := strings.Index(p.buf, "</think>")
			if idx == -1 {
				safe := safeTail(p.buf, "</think>")
				if safe > 0 {
					events = append(events, ThinkEvent{Type: "thinking", Content: p.buf[:safe]})
					p.buf = p.buf[safe:]
				}
				break
			}
			if idx > 0 {
				events = append(events, ThinkEvent{Type: "thinking", Content: p.buf[:idx]})
			}
			p.buf = p.buf[idx+len("</think>"):]
			p.inThink = false
		}
	}

	return events
}

// Reset 重置解析器状态（用于新对话）。
func (p *ThinkParser) Reset() {
	p.inThink = false
	p.buf = ""
}

// safeTail 返回 s 的安全前缀长度：末尾保留可能是 tag 前缀的部分，
// 避免 tag 被 chunk 边界截断时误发。
func safeTail(s, tag string) int {
	n := len(s)
	if n == 0 {
		return 0
	}
	for keep := 1; keep < len(tag) && keep <= n; keep++ {
		if strings.HasPrefix(tag, s[n-keep:]) {
			return n - keep
		}
	}
	return n
}
