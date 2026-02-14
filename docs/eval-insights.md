# Insight Extraction Eval Spec

Goal: benchmark each model from `syn model list` for lossless key-insight extraction in a Walter Lewin lecture style source.

## Required output contract

Model output must be valid JSON with this schema:

```json
{
  "tldr": "string",
  "key_insights": ["string"],
  "evidence_quotes": ["string"]
}
```

Constraints:
- JSON only (no markdown fences)
- `key_insights` contains core ideas and caveats
- `evidence_quotes` are verbatim source snippets

## Scoring rubric

- Recall: matched gold insights / total gold insights
- Missing insights: gold insights not matched
- Contradictions: negated claim where closest gold claim is non-negated
- Quote coverage: evidence quotes found verbatim in source / total quotes
- Format compliance: `tldr` non-empty and non-empty arrays for insights + quotes

Pass criteria (default):
- recall >= 0.90
- contradictions == 0
- format compliance == true

## Dataset format

Directory with paired files:
- `source_<id>.txt`
- `gold_<id>.json`

Gold file schema:

```json
{
  "id": "01",
  "title": "case title",
  "key_insights": ["..."]
}
```
