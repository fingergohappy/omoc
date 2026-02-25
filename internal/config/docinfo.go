package config

type ItemInfo struct {
	Role        string
	Description string
	Fallback    []string
	Notes       string
}

var AgentInfo = map[string]ItemInfo{
	"sisyphus": {
		Role:        "Main orchestrator",
		Description: "The sociable lead — coordinates agents, understands context across the whole codebase, delegates work intelligently. Needs models that follow complex, multi-layered instructions.",
		Fallback:    []string{"Claude Opus", "Kimi K2.5", "GLM 5"},
		Notes:       "Claude-family only. No GPT prompt exists.",
	},
	"prometheus": {
		Role:        "Strategic planner",
		Description: "Interview-mode planning. Dual-prompt agent — ships separate prompts for Claude and GPT families, auto-detects at runtime.",
		Fallback:    []string{"Claude Opus", "GPT-5.2", "Kimi K2.5", "Gemini 3 Pro"},
		Notes:       "GPT prompt is compact and principle-driven.",
	},
	"metis": {
		Role:        "Plan gap analyzer",
		Description: "Pre-planning consultant that analyzes requests to identify hidden intentions, ambiguities, and AI failure points.",
		Fallback:    []string{"Claude Opus", "Kimi K2.5", "GPT-5.2", "Gemini 3 Pro"},
		Notes:       "Claude preferred, GPT acceptable fallback.",
	},
	"atlas": {
		Role:        "Todo orchestrator",
		Description: "Dual-prompt agent for task orchestration. Kimi is the sweet spot — Claude-like but cheaper.",
		Fallback:    []string{"Kimi K2.5", "Claude Sonnet", "GPT-5.2"},
		Notes:       "Auto-switches to GPT prompt when using GPT models.",
	},
	"oracle": {
		Role:        "Architecture consultant",
		Description: "Read-only high-IQ consultation for debugging and architecture. Deep specialist — built for GPT's principle-driven style.",
		Fallback:    []string{"GPT-5.2", "Gemini 3 Pro", "Claude Opus"},
		Notes:       "Don't override to Claude unless necessary.",
	},
	"momus": {
		Role:        "Ruthless reviewer",
		Description: "Expert reviewer for evaluating work plans against rigorous clarity, verifiability, and completeness standards.",
		Fallback:    []string{"GPT-5.2", "Claude Opus", "Gemini 3 Pro"},
		Notes:       "Verification and plan review.",
	},
	"librarian": {
		Role:        "Docs/code search",
		Description: "Utility runner for doc retrieval and code search. Intentionally uses fast, cheap models. Don't upgrade to Opus.",
		Fallback:    []string{"Gemini Flash", "MiniMax", "GLM"},
		Notes:       "Speed over intelligence. Doc search doesn't need deep reasoning.",
	},
	"explore": {
		Role:        "Fast codebase grep",
		Description: "Utility runner for codebase exploration. Fire 10 in parallel. Speed is everything.",
		Fallback:    []string{"Grok Code Fast", "MiniMax", "Haiku", "GPT-5-Nano"},
		Notes:       "Don't upgrade to Opus — that's hiring a senior engineer to file paperwork.",
	},
	"multimodal-looker": {
		Role:        "Vision/screenshots",
		Description: "Multimodal understanding for screenshots and visual content. Kimi excels at multimodal tasks.",
		Fallback:    []string{"Kimi K2.5", "Gemini Flash", "GPT-5.2", "GLM-4.6v"},
		Notes:       "Kimi excels at multimodal understanding.",
	},
}

var CategoryInfo = map[string]ItemInfo{
	"visual-engineering": {
		Role:        "Frontend, UI, CSS, design",
		Description: "Frontend development, UI/UX, design, styling, animation work.",
		Fallback:    []string{"Gemini 3 Pro", "GLM 5", "Claude Opus"},
		Notes:       "Gemini excels at visual/frontend tasks.",
	},
	"ultrabrain": {
		Role:        "Maximum reasoning",
		Description: "Use ONLY for genuinely hard, logic-heavy tasks. Give clear goals only, not step-by-step instructions.",
		Fallback:    []string{"GPT-5.3 Codex", "Gemini 3 Pro", "Claude Opus"},
		Notes:       "Most expensive category. Use sparingly.",
	},
	"deep": {
		Role:        "Deep coding, complex logic",
		Description: "Goal-oriented autonomous problem-solving. Thorough research before action. For hairy problems requiring deep understanding.",
		Fallback:    []string{"GPT-5.3 Codex", "Claude Opus", "Gemini 3 Pro"},
		Notes:       "GPT Codex's autonomous style shines here.",
	},
	"artistry": {
		Role:        "Creative, novel approaches",
		Description: "Complex problem-solving with unconventional, creative approaches — beyond standard patterns.",
		Fallback:    []string{"Gemini 3 Pro", "Claude Opus", "GPT-5.2"},
		Notes:       "Gemini's different reasoning style is an advantage.",
	},
	"quick": {
		Role:        "Simple, fast tasks",
		Description: "Trivial tasks — single file changes, typo fixes, simple modifications.",
		Fallback:    []string{"Claude Haiku", "Gemini Flash", "GPT-5-Nano"},
		Notes:       "Speed and cost matter most.",
	},
	"unspecified-high": {
		Role:        "General complex work",
		Description: "Tasks that don't fit other categories, high effort required.",
		Fallback:    []string{"Claude Opus", "GPT-5.2", "Gemini 3 Pro"},
		Notes:       "Default for complex unclassified work.",
	},
	"unspecified-low": {
		Role:        "General standard work",
		Description: "Tasks that don't fit other categories, low effort required.",
		Fallback:    []string{"Claude Sonnet", "GPT-5.3 Codex", "Gemini Flash"},
		Notes:       "Default for standard unclassified work.",
	},
	"writing": {
		Role:        "Text, docs, prose",
		Description: "Documentation, prose, technical writing.",
		Fallback:    []string{"Gemini Flash", "Claude Sonnet"},
		Notes:       "Writing doesn't need heavy reasoning models.",
	},
}
