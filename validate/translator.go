package validate

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync"
)

//go:embed messages_en.json
var messagesEnJSON []byte

//go:embed messages_id.json
var messagesIDJSON []byte

// DefaultLang is the fallback language for unknown or unsupported langs.
const DefaultLang = "en"

// SupportedLangs lists the languages with message catalogs.
var SupportedLangs = []string{"en", "id"}

// messages[tag][lang] = template with {0}=field, {1}=param placeholders.
var (
	msgMu    sync.RWMutex
	messages = map[string]map[string]string{}
)

func init() {
	var en, id map[string]string
	if err := json.Unmarshal(messagesEnJSON, &en); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(messagesIDJSON, &id); err != nil {
		panic(err)
	}
	for tag, tmpl := range en {
		messages[tag] = map[string]string{"en": tmpl}
	}
	for tag, tmpl := range id {
		set, ok := messages[tag]
		if !ok {
			set = map[string]string{}
			messages[tag] = set
		}
		set["id"] = tmpl
	}
}

// RegisterMessage adds or overrides a message key — used both for extra
// validator tags a service registers and for business rule keys that never
// touch the validator library. Thread-safe.
func RegisterMessage(tag, enMsg, idMsg string) {
	msgMu.Lock()
	defer msgMu.Unlock()
	messages[tag] = map[string]string{"en": enMsg, "id": idMsg}
}

// normalizeLang lowercases and strips any region subtag ("ID", "id-ID" -> "id").
func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if i := strings.IndexAny(lang, "-_"); i >= 0 {
		lang = lang[:i]
	}
	return lang
}

// Translate renders the message for tag in lang, interpolating field ({0})
// and param ({1}). Unknown tags fall back to the localized generic error —
// raw tag names never leak to clients. Unknown langs fall back to English.
func Translate(tag, lang, field, param string) string {
	msgMu.RLock()
	defer msgMu.RUnlock()
	lang = normalizeLang(lang)
	set, ok := messages[tag]
	if !ok {
		set = messages["_error"]
	}
	tmpl, ok := set[lang]
	if !ok {
		tmpl = set[DefaultLang]
	}
	msg := strings.ReplaceAll(tmpl, "{0}", field)
	return strings.ReplaceAll(msg, "{1}", param)
}

// invalidInputMessage returns the localized "invalid input" fallback.
func invalidInputMessage(lang string) string {
	msgMu.RLock()
	defer msgMu.RUnlock()
	lang = normalizeLang(lang)
	if set, ok := messages["invalid_input"]; ok {
		if tmpl, ok := set[lang]; ok {
			return tmpl
		}
		return set[DefaultLang]
	}
	return "Invalid input data"
}
