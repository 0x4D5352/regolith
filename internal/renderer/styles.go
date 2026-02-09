package renderer

// Config holds all styling and dimension configuration
type Config struct {
	// Dimensions
	Padding       float64
	HorizontalGap float64
	VerticalGap   float64
	CornerRadius  float64

	// Typography
	FontFamily string
	FontSize   float64
	CharWidth  float64 // Approximate character width for text measurement

	// Colors
	BackgroundColor string
	TextColor       string
	LineColor       string
	LineWidth       float64

	// Element-specific colors
	LiteralFill          string
	CharsetFill          string
	EscapeFill           string
	AnchorFill           string
	SubexpFill           string // Used for outermost subexp (depth 0)
	SubexpStroke         string // Stroke color for subexp boxes
	SubexpColors         []string // Colors to cycle through for nested subexps (depth 1+)
	AnyCharFill          string
	FlagsFill            string
	RepeatLabelColor     string
	RecursiveRefFill     string
	CalloutFill          string
	BacktrackControlFill string
	ConditionalFill      string
}

// DefaultConfig returns the default styling configuration
func DefaultConfig() *Config {
	return &Config{
		// Dimensions
		Padding:       10,
		HorizontalGap: 10,
		VerticalGap:   5,
		CornerRadius:  3,

		// Typography
		FontFamily: "monospace",
		FontSize:   14,
		CharWidth:  8.4, // Approximate for 14px monospace

		// Colors
		BackgroundColor: "transparent",
		TextColor:       "#000",
		LineColor:       "#000",
		LineWidth:       2,

		// Element-specific colors
		LiteralFill:  "#ff6b6b", // Coral red - good contrast against grays
		CharsetFill:  "#cbcbba",
		EscapeFill:   "#bada55",
		AnchorFill:   "#6b6659",
		SubexpFill:   "none",    // Outermost subexp is transparent
		SubexpStroke: "#908c83", // Gray stroke for subexp borders
		// Accessibility-friendly colors for nested subexpressions
		// These have good contrast with black text and are distinguishable
		// for most forms of color blindness
		SubexpColors: []string{
			"#cce5ff", // Light blue
			"#d4edda", // Light green
			"#fff3cd", // Light yellow
			"#f8d7da", // Light pink
			"#e2d5f0", // Light lavender
		},
		AnyCharFill:          "#dae9e5",
		FlagsFill:            "#c8e0f9",
		RepeatLabelColor:     "#666",
		RecursiveRefFill:     "#c9b3ff", // Light lavender
		CalloutFill:          "#ffd699", // Light orange
		BacktrackControlFill: "#ffb3a7", // Light salmon
		ConditionalFill:      "#b3e5fc", // Light cyan
	}
}
