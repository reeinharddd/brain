---
name: Site Analyzer
description: Automates depth-first crawling of research targets to extract tool definitions.
---

# Site Analyzer Skill

This skill is designed to keep the brain repo updated by crawling high-value sources and extracting metadata for new MCP servers, agent skills, and automation patterns.

## Capabilities

- `parse_sources`: Reads `/home/reeinharrrd/.brain/docs/sources.md` to identify crawl targets.
- `crawl_depth`: Uses Crawl4AI to explore sub-pages and links within a target domain.
- `extract_definitions`: Performs semantic analysis on extracted Markdown to find registry-compatible definitions.

## Usage

Run the analysis via the global helper script:

```bash
/home/reeinharrrd/.brain/scripts/analyze-sources.sh
```

## Structure

- `analyze.sh`: Main entry point for the crawling loop.
- `prompts/`: System prompts for candidate extraction.
