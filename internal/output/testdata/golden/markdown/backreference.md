# Regex: `(\w+)\s+\1`

**Flavor:** JavaScript

- **Sequence**
  - **Capture group #1**
    - Escape: word
    - Quantifier: 1 or more (greedy)
  - Escape: white space
  - Quantifier: 1 or more (greedy)
  - Back-reference #1
