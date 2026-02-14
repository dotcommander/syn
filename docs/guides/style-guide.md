# Documentation Style Guide

[← Back to Guides](./)

This guide establishes consistent voice, formatting, and structure for all Syn project documentation.

## Voice and Tone

### General Principles

- **Use active voice** - Write clearly and directly
- **Address the reader as "you"** - Make documentation conversational and approachable
- **Be concise** - Get to the point without fluff
- **Use present tense** - Describe what the code does, not what it did

### Examples

```markdown
✅ GOOD
You can create a new client by calling NewClient.
The function validates your input before processing.

❌ BAD
A new client is created by the NewClient function.
The input was validated before it was processed.
```

## Code Blocks

### Language Annotations

Always specify the language for code blocks using fenced code syntax:

````markdown
```go
func Example() {}
```

```bash
npm install
```

```markdown
# Heading
```
````

Supported languages: `go`, `bash`, `markdown`, `json`, `yaml`, `toml`, `sh`

### Code Block Guidelines

- Include language annotation on every code block
- Use `bash` for shell commands and terminal output
- Use `go` for Go code snippets
- Keep examples runnable and complete
- Trim unnecessary imports from examples

## Headings

### Heading Hierarchy

Follow this heading structure:

| Level | Markdown | Usage | Example |
|-------|----------|-------|---------|
| H1 | `#` | Document title (use once) | `# Style Guide` |
| H2 | `##` | Major sections | `## Voice and Tone` |
| H3 | `###` | Subsections | `### General Principles` |
| H4 | `####` | Nested topics | `#### Examples` |

### Heading Rules

- Use **exactly one H1** per document (the title)
- Start with H2 for the first section
- Don't skip levels (H2 → H4 without H3)
- Use sentence case for headings (capitalize first word only)
- Keep headings short and descriptive
- End headings without punctuation

```markdown
✅ GOOD
# Documentation Style Guide

## Voice and Tone

### General Principles

❌ BAD
# Documentation Style Guide

# Voice and Tone  # Should be ##

### Examples  # Skipped H2 level
```

## Text Formatting

### Emphasis

- **Bold** for key terms and UI elements
- *Italic* for variables and placeholders
- `Code font` for code elements inline

### Links

- Use descriptive link text, not URLs
- Reference other docs with relative paths
- Include external links with full URLs

```markdown
✅ GOOD
See the [documentation guide](./documentation.md) for details.
Configure the `base_url` field in your config file.

❌ BAD
See https://example.com/docs for details.
Configure the base_url field.
```

## Lists

### Bullet Lists

Use for unordered items where order doesn't matter:

- Start with lowercase
- No terminal punctuation
- Keep items parallel in structure

### Numbered Lists

Use for sequential steps or ordered items:

1. Start with a capital letter
2. End with a period
3. Complete the thought

## Length and Scope

- Keep pages focused on a single topic
- Break long pages into multiple documents
- Link related concepts rather than duplicating content
- Aim for 300-800 words per page

## Review Checklist

Before publishing documentation, verify:

- [ ] Active voice is used throughout
- [ ] Reader is addressed as "you"
- [ ] All code blocks have language annotations
- [ ] Exactly one H1 heading at the top
- [ ] H2 for major sections (no skipped levels)
- [ ] Sentence case for headings
- [ ] Links use descriptive text
- [ ] Code font for inline code references
- [ ] Page is focused on a single topic

## Examples

See [Documentation Guide](./documentation.md) for well-styled documentation examples following this guide.
