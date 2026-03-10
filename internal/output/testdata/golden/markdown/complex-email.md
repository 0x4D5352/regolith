# Regex: `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`

**Flavor:** JavaScript

- **Sequence**
  - Matches one of the following, 1 or more times (greedy):
    - `a` to `z` (lowercase letters)
    - `A` to `Z` (uppercase letters)
    - `0` to `9` (digits)
    - `.`, `_`, `%`, `+`, `-` (literal characters)
  - Matches `@` literally
  - Matches one of the following, 1 or more times (greedy):
    - `a` to `z` (lowercase letters)
    - `A` to `Z` (uppercase letters)
    - `0` to `9` (digits)
    - `.`, `-` (literal characters)
  - Matches `.` literally
  - Matches one of the following, 2 or more times (greedy):
    - `a` to `z` (lowercase letters)
    - `A` to `Z` (uppercase letters)
