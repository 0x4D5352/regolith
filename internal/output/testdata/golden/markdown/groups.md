# Regex: `(foo)(?:bar)(?<name>baz)`

**Flavor:** JavaScript

- **Sequence**
  - **Capture group #1** -- captures matched text for back-reference as `\1`
    - Matches `foo` literally
  - **Non-capturing group** -- groups without capturing
    - Matches `bar` literally
  - **Named capture group #2 "name"** -- captures matched text for back-reference as `\2` or by name
    - Matches `baz` literally
