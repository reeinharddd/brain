# Memory Protocol

## Retrieval order

Use memory in this order:

1. `mem_search`
2. `mem_timeline`
3. `mem_get_observation`

This keeps context lean and only expands when the summary justifies it.

## Write protocol

For new or updated knowledge:

1. derive or confirm the project namespace
2. request or infer the topic key
3. update the existing topic when the concept already exists
4. save a concise session summary after verification

## Session summary template

```text
Project namespace:
Topic key:
Decision:
Evidence:
Validation:
Open risks:
Next step:
```
