# Regex: `(a(b(c)))`

**Flavor:** JavaScript

- **Capture group #3** -- captures matched text for back-reference as `\3`
  - **Sequence**
    - Matches `a` literally
    - **Capture group #2** -- captures matched text for back-reference as `\2`
      - **Sequence**
        - Matches `b` literally
        - **Capture group #1** -- captures matched text for back-reference as `\1`
          - Matches `c` literally
