# Regex: `foo(?=bar)(?!baz)`

**Flavor:** JavaScript

- **Sequence**
  - Matches `foo` literally
  - **Positive lookahead** -- asserts what follows matches, without consuming characters
    - Matches `bar` literally
  - **Negative lookahead** -- asserts what follows does NOT match
    - Matches `baz` literally
