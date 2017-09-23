package ankitts

type Config struct {
	SpeechApiKey string `json:"speech_bing_api_key"`
}

type Params struct {
	CollectionDir, CardType, DeckName, LanguageLocale, SpeechColumnsStr string
}

