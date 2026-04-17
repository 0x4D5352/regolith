package posix_bre

import "github.com/0x4d5352/regolith/internal/flavor/helpers"

// parseInt is referenced by the generated parser; delegate to shared.
func parseInt(v any) int { return helpers.ParseInt(v) }
