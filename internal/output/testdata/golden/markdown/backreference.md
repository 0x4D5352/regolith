# Regex: `(\w+)\s+\1`

**Flavor:** JavaScript

- **Sequence**
  - **Capture group #1** -- captures matched text for back-reference as `\1`
    - Matches any word character `\w` (a-z, A-Z, 0-9, _), 1 or more times (greedy)
  - Matches any whitespace character `\s` (space, tab, newline, etc.), 1 or more times (greedy)
  - Matches the same text previously captured by group #1
