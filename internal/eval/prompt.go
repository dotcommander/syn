package eval

import "fmt"

// BuildPrompt builds a strict JSON extraction prompt.
func BuildPrompt(source string) string {
	return fmt.Sprintf(`You are evaluating key-insight extraction quality.

Task:
1) Produce a short TL;DR.
2) Extract key insights from the source without losing critical meaning.
3) Provide direct evidence quotes copied verbatim from the source.

Rules:
- Return JSON only.
- Do not include markdown fences.
- Keep claims faithful to the source.
- Include the most important concepts and caveats.

Return schema:
{
  "tldr": "string",
  "key_insights": ["string"],
  "evidence_quotes": ["string"]
}

Source:
%s`, source)
}
