# Regex: `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`

**Flavor:** JavaScript

- **Sequence**
  - **Character class**: one of
    - Range `a` to `z`
    - Range `A` to `Z`
    - Range `0` to `9`
    - `.`
    - `_`
    - `%`
    - `+`
    - `-`
  - Quantifier: 1 or more (greedy)
  - Literal `@`
  - **Character class**: one of
    - Range `a` to `z`
    - Range `A` to `Z`
    - Range `0` to `9`
    - `.`
    - `-`
  - Quantifier: 1 or more (greedy)
  - Literal `.`
  - **Character class**: one of
    - Range `a` to `z`
    - Range `A` to `Z`
  - Quantifier: 2 or more (greedy)
