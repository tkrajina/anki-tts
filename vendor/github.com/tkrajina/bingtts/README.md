# Go Bing TTS
Use TTS voices available on Microsoft's Cognitive Services.

# Installation

`go get -u github.com/stevemurr/bingtts`

#### Available Locales and Voices

```
var voices = map[string]string{
	"ar-eg female": "Microsoft Server Speech Text to Speech Voice (ar-EG, Hoda)",
	"de-de female": "Microsoft Server Speech Text to Speech Voice (de-DE, Hedda)",
	"de-de male":   "Microsoft Server Speech Text to Speech Voice (de-DE, Stefan, Apollo)",
	"en-au female": "Microsoft Server Speech Text to Speech Voice (en-AU, Catherine)",
	"en-ca female": "Microsoft Server Speech Text to Speech Voice (en-CA, Linda)",
	"en-gb female": "Microsoft Server Speech Text to Speech Voice (en-GB, Susan, Apollo)",
	"en-gb male":   "Microsoft Server Speech Text to Speech Voice (en-GB, George, Apollo)",
	"en-in male":   "Microsoft Server Speech Text to Speech Voice (en-IN, Ravi, Apollo)",
	"en-us female": "Microsoft Server Speech Text to Speech Voice (en-US, ZiraRUS)",
	"en-us male":   "Microsoft Server Speech Text to Speech Voice (en-US, BenjaminRUS)",
	"es-es female": "Microsoft Server Speech Text to Speech Voice (es-ES, Laura, Apollo)",
	"es-es male":   "Microsoft Server Speech Text to Speech Voice (es-ES, Pablo, Apollo)",
	"es-mx male":   "Microsoft Server Speech Text to Speech Voice (es-MX, Raul, Apollo)",
	"fr-ca female": "Microsoft Server Speech Text to Speech Voice (fr-CA, Caroline)",
	"fr-fr female": "Microsoft Server Speech Text to Speech Voice (fr-FR, Julie, Apollo)",
	"fr-fr male":   "Microsoft Server Speech Text to Speech Voice (fr-FR, Paul, Apollo)",
	"it-it male":   "Microsoft Server Speech Text to Speech Voice (it-IT, Cosimo, Apollo)",
	"ja-jp female": "Microsoft Server Speech Text to Speech Voice (ja-JP, Ayumi, Apollo)",
	"ja-jp male":   "Microsoft Server Speech Text to Speech Voice (ja-JP, Ichiro, Apollo)",
	"pt-br male":   "Microsoft Server Speech Text to Speech Voice (pt-BR, Daniel, Apollo)",
	"ru-ru female": "Microsoft Server Speech Text to Speech Voice (ru-RU, Irina, Apollo)",
	"ru-ru male":   "Microsoft Server Speech Text to Speech Voice (ru-RU, Pavel, Apollo)",
	"zh-cn female": "Microsoft Server Speech Text to Speech Voice (zh-CN, Yaoyao, Apollo)",
	"zh-cn male":   "Microsoft Server Speech Text to Speech Voice (zh-CN, Kangkang, Apollo)",
	"zh-hk female": "Microsoft Server Speech Text to Speech Voice (zh-HK, Tracy, Apollo)",
	"zh-hk male":   "Microsoft Server Speech Text to Speech Voice (zh-HK, Danny, Apollo)",
	"zh-tw female": "Microsoft Server Speech Text to Speech Voice (zh-TW, Yating, Apollo)",
	"zh-tw male":   "Microsoft Server Speech Text to Speech Voice (zh-TW, Zhiwei, Apollo)",
}
```
# Example
```
package main

import (
    "log"
    "io/ioutil"

    "github.com/stevemurr/bingtts"
)
func main() {
    // See what voices you can use
    for key, value := range bingtts.GetVoices() {
        log.Printf("%s: %s", key, value)
    }
    // Pass in your key from https://www.microsoft.com/cognitive-services/en-us/sign-up
    // Key lasts 10 minutes so if you're doing serious synthesis, write some code to reuse the key
    key := "YOUR KEY HERE"
    token, err := bingtts.IssueToken(key)
    if err != nil {
        log.Println(err)
    }
    // Synthesize
    res, err := bingtts.Synthesize(
        token,
        "oh my god did you hear?  trump is dead!",
        "es-mx",
        "male",
        bingtts.RIFF16Bit16kHzMonoPCM)
    if err != nil {
        log.Println(err)
    }
    // res is []byte
    ioutil.WriteFile("test.wav", res, 0644)
}
```
