package ankitts

import (
	"bytes"
	"fmt"
	"github.com/tkrajina/bingtts"
	"io/ioutil"
	"strings"
	"unicode"
)

type Gender string

const (
	Make   Gender = "male"
	Female Gender = "female"
)

func Retrieve(params Params, config Config, gender Gender, text string, targetDir, destFilename string) error {
	token, err := bingtts.IssueToken(config.SpeechApiKey)
	if err != nil {
		return err
	}

	fmt.Println("token=", token)

	// Synthesize
	res, err := bingtts.Synthesize(
		token,
		text,
		params.LanguageLocale,
		bingtts.Gender(gender),
		"",
		bingtts.Audio16khz32kbitrateMonoMp3)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(destFilename, res, 0644); err != nil {
		return err
	}

	return nil
}

func PrepareDestfilename(text string) string {
	var res bytes.Buffer
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			res.WriteRune(r)
		} else {
			res.WriteRune('_')
		}
	}
	return strings.Trim(res.String(), "_")
}
