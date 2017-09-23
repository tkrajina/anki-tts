package bingtts

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Voice struct {
	Locale, Description, VoiceName string
	Gender                         Gender
}

var voicesStr = `ar-EG*	Female	"Microsoft Server Speech Text to Speech Voice (ar-EG, Hoda)"
ar-SA	Male	"Microsoft Server Speech Text to Speech Voice (ar-SA, Naayf)"
ca-ES	Female	"Microsoft Server Speech Text to Speech Voice (ca-ES, HerenaRUS)"
cs-CZ	Male	"Microsoft Server Speech Text to Speech Voice (cs-CZ, Vit)"
da-DK	Female	"Microsoft Server Speech Text to Speech Voice (da-DK, HelleRUS)"
de-AT	Male	"Microsoft Server Speech Text to Speech Voice (de-AT, Michael)"
de-CH	Male	"Microsoft Server Speech Text to Speech Voice (de-CH, Karsten)"
de-DE	Female	"Microsoft Server Speech Text to Speech Voice (de-DE, Hedda) "
de-DE	Female	"Microsoft Server Speech Text to Speech Voice (de-DE, HeddaRUS)"
de-DE	Male	"Microsoft Server Speech Text to Speech Voice (de-DE, Stefan, Apollo) "
el-GR	Male	"Microsoft Server Speech Text to Speech Voice (el-GR, Stefanos)"
en-AU	Female	"Microsoft Server Speech Text to Speech Voice (en-AU, Catherine) "
en-AU	Female	"Microsoft Server Speech Text to Speech Voice (en-AU, HayleyRUS)"
en-CA	Female	"Microsoft Server Speech Text to Speech Voice (en-CA, Linda)"
en-CA	Female	"Microsoft Server Speech Text to Speech Voice (en-CA, HeatherRUS)"
en-GB	Female	"Microsoft Server Speech Text to Speech Voice (en-GB, Susan, Apollo)"
en-GB	Female	"Microsoft Server Speech Text to Speech Voice (en-GB, HazelRUS)"
en-GB	Male	"Microsoft Server Speech Text to Speech Voice (en-GB, George, Apollo)"
en-IE	Male	"Microsoft Server Speech Text to Speech Voice (en-IE, Shaun)"
en-IN	Female	"Microsoft Server Speech Text to Speech Voice (en-IN, Heera, Apollo)"
en-IN	Female	"Microsoft Server Speech Text to Speech Voice (en-IN, PriyaRUS)"
en-IN	Male	"Microsoft Server Speech Text to Speech Voice (en-IN, Ravi, Apollo) "
en-US	Female	"Microsoft Server Speech Text to Speech Voice (en-US, ZiraRUS)"
en-US	Female	"Microsoft Server Speech Text to Speech Voice (en-US, JessaRUS)"
en-US	Male	"Microsoft Server Speech Text to Speech Voice (en-US, BenjaminRUS)"
es-ES	Female	"Microsoft Server Speech Text to Speech Voice (es-ES, Laura, Apollo)"
es-ES	Female	"Microsoft Server Speech Text to Speech Voice (es-ES, HelenaRUS)"
es-ES	Male	"Microsoft Server Speech Text to Speech Voice (es-ES, Pablo, Apollo)"
es-MX	Female	"Microsoft Server Speech Text to Speech Voice (es-MX, HildaRUS)"
es-MX	Male	"Microsoft Server Speech Text to Speech Voice (es-MX, Raul, Apollo)"
fi-FI	Female	"Microsoft Server Speech Text to Speech Voice (fi-FI, HeidiRUS)"
fr-CA	Female	"Microsoft Server Speech Text to Speech Voice (fr-CA, Caroline)"
fr-CA	Female	"Microsoft Server Speech Text to Speech Voice (fr-CA, HarmonieRUS)"
fr-CH	Male	"Microsoft Server Speech Text to Speech Voice (fr-CH, Guillaume)"
fr-FR	Female	"Microsoft Server Speech Text to Speech Voice (fr-FR, Julie, Apollo)"
fr-FR	Female	"Microsoft Server Speech Text to Speech Voice (fr-FR, HortenseRUS)"
fr-FR	Male	"Microsoft Server Speech Text to Speech Voice (fr-FR, Paul, Apollo)"
he-IL	Male	"Microsoft Server Speech Text to Speech Voice (he-IL, Asaf)"
hi-IN	Female	"Microsoft Server Speech Text to Speech Voice (hi-IN, Kalpana, Apollo)"
hi-IN	Female	"Microsoft Server Speech Text to Speech Voice (hi-IN, Kalpana)"
hi-IN	Male	"Microsoft Server Speech Text to Speech Voice (hi-IN, Hemant)"
hu-HU	Male	"Microsoft Server Speech Text to Speech Voice (hu-HU, Szabolcs)"
id-ID	Male	"Microsoft Server Speech Text to Speech Voice (id-ID, Andika)"
it-IT	Male	"Microsoft Server Speech Text to Speech Voice (it-IT, Cosimo, Apollo)"
ja-JP	Female	"Microsoft Server Speech Text to Speech Voice (ja-JP, Ayumi, Apollo)"
ja-JP	Male	"Microsoft Server Speech Text to Speech Voice (ja-JP, Ichiro, Apollo)"
ja-JP	Female	"Microsoft Server Speech Text to Speech Voice (ja-JP, HarukaRUS)"
ja-JP	Female	"Microsoft Server Speech Text to Speech Voice (ja-JP, LuciaRUS)"
ja-JP	Male	"Microsoft Server Speech Text to Speech Voice (ja-JP, EkaterinaRUS)"
ko-KR	Female	"Microsoft Server Speech Text to Speech Voice (ko-KR, HeamiRUS)"
nb-NO	Female	"Microsoft Server Speech Text to Speech Voice (nb-NO, HuldaRUS)"
nl-NL	Female	"Microsoft Server Speech Text to Speech Voice (nl-NL, HannaRUS)"
pl-PL	Female	"Microsoft Server Speech Text to Speech Voice (pl-PL, PaulinaRUS)"
pt-BR	Female	"Microsoft Server Speech Text to Speech Voice (pt-BR, HeloisaRUS)"
pt-BR	Male	"Microsoft Server Speech Text to Speech Voice (pt-BR, Daniel, Apollo)"
pt-PT	Female	"Microsoft Server Speech Text to Speech Voice (pt-PT, HeliaRUS)"
ro-RO	Male	"Microsoft Server Speech Text to Speech Voice (ro-RO, Andrei)"
ru-RU	Female	"Microsoft Server Speech Text to Speech Voice (ru-RU, Irina, Apollo)"
ru-RU	Male	"Microsoft Server Speech Text to Speech Voice (ru-RU, Pavel, Apollo)"
sk-SK	Male	"Microsoft Server Speech Text to Speech Voice (sk-SK, Filip)"
sv-SE	Female	"Microsoft Server Speech Text to Speech Voice (sv-SE, HedvigRUS)"
th-TH	Male	"Microsoft Server Speech Text to Speech Voice (th-TH, Pattara)"
tr-TR	Female	"Microsoft Server Speech Text to Speech Voice (tr-TR, SedaRUS)"
zh-CN	Female	"Microsoft Server Speech Text to Speech Voice (zh-CN, HuihuiRUS)"
zh-CN	Female	"Microsoft Server Speech Text to Speech Voice (zh-CN, Yaoyao, Apollo)"
zh-CN	Male	"Microsoft Server Speech Text to Speech Voice (zh-CN, Kangkang, Apollo)"
zh-HK	Female	"Microsoft Server Speech Text to Speech Voice (zh-HK, Tracy, Apollo)"
zh-HK	Female	"Microsoft Server Speech Text to Speech Voice (zh-HK, TracyRUS)"
zh-HK	Male	"Microsoft Server Speech Text to Speech Voice (zh-HK, Danny, Apollo)"
zh-TW	Female	"Microsoft Server Speech Text to Speech Voice (zh-TW, Yating, Apollo)"
zh-TW	Female	"Microsoft Server Speech Text to Speech Voice (zh-TW, HanHanRUS)"
zh-TW	Male	"Microsoft Server Speech Text to Speech Voice (zh-TW, Zhiwei, Apollo)"`

var voices map[string][]Voice
var voiceNameRegexp = regexp.MustCompile(`\(.*?,(.*?)\)`)

type Gender string

const (
	Male   Gender = "male"
	Female        = "female"
)

func init() {
	voices = map[string][]Voice{}
	for _, line := range strings.Split(voicesStr, "\n") {
		parts := strings.Split(line, "\t")
		locale := strings.Replace(parts[0], "*", "", 1)
		gender := strings.ToLower(parts[1])
		description := parts[2]
		voiceName := voiceNameRegexp.FindAllStringSubmatch(description, -1)[0][1]

		voiceKey := strings.ToLower(fmt.Sprintf("%s %s", locale, gender))
		v := Voice{
			Locale:      locale,
			VoiceName:   strings.TrimSpace(voiceName),
			Gender:      Gender(gender),
			Description: strings.Trim(description, `" `),
		}
		voices[voiceKey] = append(voices[voiceKey], v)
	}
}

const (
	bingSpeechTokenEndpoint = "https://api.cognitive.microsoft.com/sts/v1.0/issueToken"
	bingSpeechEndpointTTS   = "https://speech.platform.bing.com/synthesize"
)

type OutputType string

const (
	RIFF8Bit8kHzMonoPCM          OutputType = "riff-8khz-8bit-mono-mulaw"
	RIFF16Bit16kHzMonoPCM                   = "riff-16khz-16bit-mono-pcm"
	RIFF16khz16kbpsMonoSiren                = "riff-16khz-16kbps-mono-siren"
	RAW8Bit8kHzMonoMulaw                    = "raw-8khz-8bit-mono-mulaw"
	RAW16Bit16kHzMonoMulaw                  = "raw-16khz-16bit-mono-pcm"
	Ssml16khz16bitMonoTts                   = "ssml-16khz-16bit-mono-tts"
	Audio16khz16kbpsMonoSiren               = "audio-16khz-16kbps-mono-siren"
	Audio16khz128kbitrateMonoMp3            = "audio-16khz-128kbitrate-mono-mp3"
	Audio16khz64kbitrateMonoMp3             = "audio-16khz-64kbitrate-mono-mp3"
	Audio16khz32kbitrateMonoMp3             = "audio-16khz-32kbitrate-mono-mp3"
)

func getSSML(locale string, v Voice, gender Gender, text string) string {
	return fmt.Sprintf(`<speak version='1.0' xml:lang='%s'><voice name='%s' xml:lang='%s' xml:gender='%s'>%s</voice></speak>`,
		locale,
		v.Description,
		locale,
		gender,
		text)
}

// Synthesize --
func Synthesize(token, text, locale string, gender Gender, voiceName string, outputFormat OutputType) ([]byte, error) {
	client := &http.Client{}
	voices, found := voices[fmt.Sprintf("%s %s", strings.ToLower(locale), strings.ToLower(string(gender)))]
	if !found {
		return nil, fmt.Errorf("No voice for %s %s", locale, gender)
	}

	voiceName = strings.TrimSpace(voiceName)
	var voice *Voice
	var voiceNames []string
	for _, v := range voices {
		voiceNames = append(voiceNames, v.VoiceName)
		if strings.Contains(strings.ToLower(v.VoiceName), strings.ToLower(voiceName)) {
			voice = &v
			break
		}
	}

	if voice == nil {
		return nil, fmt.Errorf("No voice for found for %s, available voice names: %s", voiceName, strings.Join(voiceNames, ", "))
	}

	ssml := getSSML(locale, voices[0], gender, text)
	req, err := http.NewRequest("POST", bingSpeechEndpointTTS, bytes.NewBufferString(ssml))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Length", strconv.Itoa(len(token)))
	req.Header.Add("Content-Type", "application/ssml+xml")
	req.Header.Add("X-Microsoft-OutputFormat", string(outputFormat))
	req.Header.Add("X-Search-AppId", "00000000000000000000000000000000")
	req.Header.Add("X-Search-ClientID", "00000000000000000000000000000000")
	req.Header.Add("User-Agent", "go-bing-tts")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", res.Status)
	}
	defer res.Body.Close()
	size, err := strconv.Atoi(res.Header.Get("Content-Length"))
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, size))
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetVoices -- Return voices availble on cognitive services
func GetVoices() map[string][]Voice {
	return voices
}

// IssueToken -- Get a JWT token from cognitive services
func IssueToken(key string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", bingSpeechTokenEndpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Ocp-Apim-Subscription-Key", key)
	req.Header.Add("Content-Length", "0")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s", res.Status)
	}

	defer res.Body.Close()
	size, err := strconv.Atoi(res.Header.Get("Content-Length"))
	if err != nil {
		return "", err
	}
	buf := make([]byte, size)
	res.Body.Read(buf)
	return string(buf), nil
}
