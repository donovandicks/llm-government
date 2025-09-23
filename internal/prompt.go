package internal

import (
	"fmt"
	"strings"
)

func writePrompt(builder *strings.Builder, tag, content string) {
	fmt.Fprintf(builder, "<%s>%s</%s>\n", tag, content, tag)
}

type Example struct {
	Input  string
	Output string
}

type TaskOption func(builder *strings.Builder)

func WithItems(items ...string) TaskOption {
	return func(builder *strings.Builder) {
		builder.WriteString("<list>\n")
		for _, item := range items {
			writePrompt(builder, "item", item)
		}
		builder.WriteString("</list>\n")
	}
}

func WithOutputFormat(format string) TaskOption {
	return func(builder *strings.Builder) {
		writePrompt(builder, "output-format", format)
	}
}

// PromptBuilder is a tool for programmatically constructing LLM prompts
// using the [POML](https://microsoft.github.io/poml/latest/) language.
//
// The builder is designed to add convenience to writing prompts by hand
// as well as static checking and correctness.
type PromptBuilder struct {
	builder strings.Builder
}

func (p *PromptBuilder) WithRole(role string) *PromptBuilder {
	writePrompt(&p.builder, "role", role)
	return p
}

func (p *PromptBuilder) WithSystemMessage(msg string) *PromptBuilder {
	writePrompt(&p.builder, "system-msg", msg)
	return p
}

func (p *PromptBuilder) WithUserMessage(msg string) *PromptBuilder {
	writePrompt(&p.builder, "user-msg", msg)
	return p
}

func (p *PromptBuilder) WithIntroducer(msg string) *PromptBuilder {
	writePrompt(&p.builder, "introducer", msg)
	return p
}

func (p *PromptBuilder) WithTask(task string, opts ...TaskOption) *PromptBuilder {
	fmt.Fprintf(&p.builder, "<task>%s\n", task)
	for _, opt := range opts {
		opt(&p.builder)
	}
	fmt.Fprintf(&p.builder, "</task>\n")
	return p
}

func (p *PromptBuilder) WithItems(items ...string) *PromptBuilder {
	WithItems(items...)(&p.builder)
	return p
}

func (p *PromptBuilder) WithExample(input, output string) *PromptBuilder {
	fmt.Fprintf(&p.builder, "<example>\n")
	writePrompt(&p.builder, "input", input)
	writePrompt(&p.builder, "output", output)
	fmt.Fprintf(&p.builder, "</example>\n")
	return p
}

func (p *PromptBuilder) WithExamples(exs []Example) *PromptBuilder {
	fmt.Fprintf(&p.builder, "<examples>\n")
	for _, e := range exs {
		p.WithExample(e.Input, e.Output)
	}
	fmt.Fprintf(&p.builder, "</examples>")
	return p
}

func (p *PromptBuilder) WithCode(code, language string) *PromptBuilder {
	fmt.Fprintf(&p.builder, "<code language=\"%s\">\n%s\n</code>\n", language, code)
	return p
}

func (p *PromptBuilder) WithParagraph(text string) *PromptBuilder {
	writePrompt(&p.builder, "p", text)
	return p
}

func (p *PromptBuilder) Build() string {
	output := p.builder.String()
	p.builder.Reset()
	return output
}
