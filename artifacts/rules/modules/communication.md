## Module: Communication

### How I communicate with AI agents

**Context first**: Always open with context before the request.
Bad: "Fix the login bug"
Good: "The login endpoint at POST /auth/login returns 500 when the email contains uppercase letters. Fix it without changing the response structure."

**Be specific about scope**: Clearly state what's in and out of scope.
Bad: "Improve this code"  
Good: "Improve the readability of this function. Don't change the logic or the function signature."

**Ask for reasoning when uncertain**: If I don't understand why an agent chose an approach, I ask.
"Why did you choose approach X over approach Y?"

**Acknowledge mistakes openly**: If I gave wrong context, I correct it immediately.
Don't waste tokens on retries with the same bad context.

### How AI agents should communicate with me

- **Be direct**: Skip preamble. Don't start responses with "Sure!" or "Great question!"
- **No Emojis or Symbols**: NEVER use emojis (😀, 🚀, etc.) or decorative symbols ([PASS], ->, [FAIL]). Use plain text.
- **Show your work briefly**: When making significant decisions, explain the tradeoff in 1-2 sentences
- **Use lists for steps**: Sequential tasks should be numbered. Options should be bulleted
- **Flag uncertainty**: If you're not sure, say so. Don't hallucinate confidence
- **Ask before assuming**: If a requirement is ambiguous, ask ONE clarifying question before proceeding
- **Format code correctly**: Use proper code blocks with language tags
- **Cite sources when relevant**: If referencing a library or pattern, mention where it's documented

### Response length guidelines


- Simple questions -> 1-3 sentences
- Code tasks -> Code + brief explanation only
- Architecture/planning -> Structured with headers, as long as needed
- Never pad responses. Quality > quantity.
