package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode"

	"bitbucket.org/puzz/anki-tts/ankitts"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tkrajina/anki"
)

var (
	config         ankitts.Config
	params         ankitts.Params
	askedForUpdate bool
	update         bool
	speechColumns  map[string]bool
)

func init() {
	flag.StringVar(&params.CollectionDir, "c", "", "Collection directory")
	flag.StringVar(&params.CardType, "t", "", "Collection type")
	flag.StringVar(&params.DeckName, "d", "", "Deck name")
	flag.StringVar(&params.LanguageLocale, "l", "", "Locale")
	flag.StringVar(&params.SpeechColumnsStr, "s", "Back", "Spech columns (coma delimited)")

	flag.Parse()
	fmt.Println("Init")

	fmt.Printf("params=%#v\n", params)

	if params.CollectionDir == "" || params.CardType == "" || params.DeckName == "" || params.SpeechColumnsStr == "" || params.LanguageLocale == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	speechColumns = map[string]bool{}
	for _, speechColumn := range strings.Split(params.SpeechColumnsStr, ",") {
		speechColumns[strings.TrimSpace(speechColumn)] = true
	}

	usr, err := user.Current()
	panicIfErrf(err, "getting user")
	cfgFile := path.Join(usr.HomeDir, ".anki-tts")

	cfgBytes, err := ioutil.ReadFile(cfgFile)
	panicIfErrf(err, "reading %s", cfgFile)

	err = json.Unmarshal(cfgBytes, &config)
	panicIfErrf(err, "unmarshalling %s", string(cfgBytes))
}

func main() {
	backup()

	collectionsDb := path.Join(params.CollectionDir, "collection.anki2")
	db, err := anki.OpenOriginalDB(collectionsDb)
	panicIfErrf(err, "opening db %s", collectionsDb)

	defer func() {
		fmt.Println("Closing db")
		db.Close()
	}()

	collection, err := db.Collection()
	panicIfErrf(err, "getting collection")

	for modelId, model := range collection.Models {
		fmt.Printf("model [%d] %s deck=%d\n", modelId, model.Name, model.DeckID)
	}

	for deckId, deck := range collection.Decks {
		fmt.Printf("deck [%d/%d] %s\n", deckId, deck.ID, deck.Name)
	}

	notesById := map[anki.ID]anki.Note{}
	notes, err := db.Notes()
	panicIfErrf(err, "getting notes")
	for notes.Next() {
		note, err := notes.Note()
		panicIfErrf(err, "getting note")
		notesById[note.ID] = *note
	}
	notes.Close()

	allCards := []anki.Card{}

	cards, err := db.Cards()
	panicIfErrf(err, "getting cards")
	for cards.Next() {
		card, err := cards.Card()
		panicIfErrf(err, "getting card")
		allCards = append(allCards, *card)
	}
	cards.Close()

	for _, card := range allCards {
		deck, found := collection.Decks[card.DeckID]
		//fmt.Println("found", deck.Name, deckName)
		if found && deck.Name == params.DeckName {
			note, found := notesById[card.NoteID]
			if !found {
				fmt.Printf("Note %d not found\n", card.NoteID)
				continue
			}

			model, found := collection.Models[note.ModelID]
			if !found {
				fmt.Println("Model not found", note.ModelID, note.FieldValues)
				continue
			}

			//fmt.Println("*", model.Name, cardType)
			if model.Name == params.CardType {
				process(db, note, *model)
			} else {
				fmt.Printf("Note %#v in deck %s but not of type %s\n", note.FieldValues, params.DeckName, params.CardType)
			}
		}
	}
}

func backup() {
	backupFilename := fmt.Sprintf("%s-%s.tar", path.Base(params.CollectionDir), time.Now().Format(time.RFC3339))
	cmd := exec.Command("tar", "-cvf", backupFilename, params.CollectionDir)
	bytes, err := cmd.CombinedOutput()
	panicIfErrf(err, "backup: %s", string(bytes))
	fmt.Println(string(bytes), "\n")

	fmt.Println("Backup:", backupFilename)
}

func process(db *anki.DB, note anki.Note, model anki.Model) {
	mediaDir := path.Join(params.CollectionDir, "collection.media")

	for n := range model.Fields {
		fieldName := model.Fields[n].Name
		if _, found := speechColumns[fieldName]; found {
			text := strings.TrimSpace(note.FieldValues[n])
			original := text
			text = regexp.MustCompile(`\[.*?\]`).ReplaceAllString(text, "")
			if len(text) > 0 {
				//if !strings.Contains(text, "[sound:") {
				fmt.Printf("field %s=%s\n", fieldName, text)
				//}
				speechFile := fmt.Sprintf("%s/%s-%s.mp3", mediaDir, params.LanguageLocale, ankitts.PrepareDestfilename(text))
				note.FieldValues[n] = text + fmt.Sprintf("[sound:%s]", path.Base(speechFile))

				if original == note.FieldValues[n] {
					fmt.Printf("unchanged %s -> %s\n", original, note.FieldValues[n])
				} else {
					err := ankitts.Retrieve(params, config, ankitts.Female, prepareText(text), mediaDir, speechFile)
					panicIfErrf(err, "retrieving speech file")

					fmt.Printf("changed %s -> %s\n", original, note.FieldValues[n])
					if !askedForUpdate {
						askedForUpdate = true
						fmt.Println("Update? [y/n]")
						var answer string
						fmt.Scan(&answer)
						update = answer == "y"
					}
					if update {
						fieldsJoined := strings.Join(note.FieldValues, anki.FieldValuesDelimiter)
						update := "update notes set flds=?, mod=?, usn=-1 where id=?"
						updateParams := []interface{}{fieldsJoined, int(time.Now().Unix() / 1000), note.ID}
						fmt.Printf("sql: %s with params %#v\n", update, updateParams)

						res, err := db.Exec(update, updateParams...)
						panicIfErrf(err, "updatind %d, fields %s", note.ID, fieldsJoined)
						fmt.Println("Updated", note.ID, "to", fieldsJoined)

						affected, err := res.RowsAffected()
						panicIfErrf(err, "getting affected rows %d", note.ID)
						if affected != 1 {
							panic("Err updating")
						}
					}
				}
			}
		}
	}
}

var ignoreTextForSpechRegexp = regexp.MustCompile(`\[.*?\]`)

func prepareText(str string) string {

	str = ignoreTextForSpechRegexp.ReplaceAllString(str, "")

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(str)))
	panicIfErrf(err, "removing html from %s", str)
	str = doc.Text()

	res := ""
	for _, r := range str {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			res += string(r)
		} else if r == '.' || r == ',' || r == '!' || r == '?' {
			res += string(r)
		} else {
			res += " "
		}
	}
	return regexp.MustCompile(`\s+`).ReplaceAllString(res, " ")
}

func panicIfErrf(err error, msgf string, args ...interface{}) {
	if err == nil {
		return
	}
	panic(fmt.Sprintf("%s: %s", fmt.Sprintf(msgf, args...), err.Error()))
}
