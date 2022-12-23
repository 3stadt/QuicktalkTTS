package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"

	"fmt"
	"os"
)

type bc struct {
	label string
	text  string
}

type he struct {
	hash string
	text string
}

type ttsTextField struct {
	widget.Entry
}

var sess *session.Session
var svc *polly.Polly
var list *widget.List
var histFileFolder = "history"
var voiceName = "Daniel"
var history = []he{}
var permHistory = map[string]struct{}{}
var buttonContent = []bc{
	{"Hi", "Moin Moin!"},
	{"Bye", "Tschüss!"},
	{"Ok", "Ok!"},
	{"Wb", "Willkommen zurück!"},
	{"brb", "Bin gleich wieder zurück."},
	{"re", "Bin wieder da."},
	{"thx", "Dankeschön!"},
	{"Ja", "Ja!"},
	{"Nein", "Nein."},
	{"Vielleicht", "Vielleicht!"},
}

func init() {

	sess = session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String("eu-central-1")},
		SharedConfigState: session.SharedConfigEnable,
		SharedConfigFiles: []string{".aws_credentials"},
	}))
	svc = polly.New(sess)

	if len(os.Args) > 1 {
		if os.Args[1] == "-f" {
			voiceName = "Vicki"
		}
	}
}

func main() {

	ttsGui := app.New()
	w := ttsGui.NewWindow("QuickTalkTTS")
	w.Resize(fyne.NewSize(500, 800))

	input := newTtsTextField()
	input.SetPlaceHolder("Text eingeben...")

	howto := widget.NewLabel("Text eingeben, Enter drücken zum Vorlesen.")

	list = widget.NewList(
		func() int {
			return len(history)
		},
		func() fyne.CanvasObject {
			return widget.NewButton("...", func() {})
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			text := history[i].text
			o.(*widget.Button).Text = text
			o.(*widget.Button).OnTapped = func() {
				createAndPlay(text, false)
			}
		})

	buttons := container.NewGridWithColumns(4)

	for _, entry := range buttonContent {
		text := entry.text
		buttons.Add(widget.NewButton(entry.label, func() {
			createAndPlay(text, true)
		}))
		permHistory[createMd5Hash(text)] = struct{}{}
	}

	//content := container.NewVBox(howto, input, buttons, list)
	menu := container.NewVBox(howto, input, buttons)
	content := container.NewBorder(menu, nil, nil, nil, list)

	w.SetContent(content)

	w.Show()
	ttsGui.Run()
	cleanup()
}

func (m *ttsTextField) TypedKey(key *fyne.KeyEvent) {
	if key.Name == "Return" {
		m.Entry.Disable()

		createAndPlay(m.Text, false)

		m.Entry.Text = ""
		m.Entry.Enable()
	} else {
		m.Entry.TypedKey(key)
	}
}

func createAndPlay(text string, keepForever bool) {
	hash := createMd5Hash(text)
	preRecFileName := fmt.Sprintf("%s/%s.mp3", histFileFolder, hash)

	if !fileExists(preRecFileName) {
		createMP3(text, preRecFileName)
	}

	if !keepForever {
		addToHistory(hash, text)
	} else {
		permHistory[hash] = struct{}{}
	}

	play(preRecFileName)
}

func addToHistory(hash string, text string) {
	for _, entry := range history {
		if entry.hash == hash {
			return
		}
	}
	history = append(history, he{hash, text})
	list.Refresh()
}

func newTtsTextField() *ttsTextField {
	entry := &ttsTextField{}
	entry.ExtendBaseWidget(entry)

	return entry
}

func createMP3(text, fileName string) {
	input := &polly.SynthesizeSpeechInput{
		OutputFormat: aws.String("mp3"),
		Text:         aws.String(text),
		VoiceId:      aws.String(voiceName),
		Engine:       aws.String("neural"),
	}

	output, err := svc.SynthesizeSpeech(input)
	if err != nil {
		fmt.Println("Got error calling SynthesizeSpeech:")
		fmt.Print(err.Error())
		os.Exit(1)
	}

	names := strings.Split(fileName, ".")
	name := names[0]
	mp3File := name + ".mp3"

	outFile, err := os.Create(mp3File)
	if err != nil {
		fmt.Println("Got error creating " + mp3File + ":")
		fmt.Print(err.Error())
		os.Exit(1)
	}

	defer outFile.Close()
	_, err = io.Copy(outFile, output.AudioStream)
	if err != nil {
		fmt.Println("Got error saving MP3:")
		fmt.Print(err.Error())
		os.Exit(1)
	}
}

func play(fileName string) {

	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Got error opening MP3:")
		fmt.Print(err.Error())
		os.Exit(1)
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		fmt.Println("Got error creating decoder:")
		fmt.Print(err.Error())
		os.Exit(1)
	}

	c, ready, err := oto.NewContext(d.SampleRate(), 2, 2)
	if err != nil {
		fmt.Println("Got error creating context:")
		fmt.Print(err.Error())
		os.Exit(1)
	}
	<-ready

	p := c.NewPlayer(d)
	defer p.Close()
	p.Play()

	for {
		time.Sleep(time.Second)
		if !p.IsPlaying() {
			break
		}
	}
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func createMd5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func cleanup() {
	dir, err := os.ReadDir(histFileFolder)
	if err != nil {
		log.Fatalln(err)
	}

	dLen := len(dir)

	for i := 0; i < dLen; i++ {
		if !strings.HasSuffix(dir[i].Name(), ".mp3") || dir[i].IsDir() {
			continue
		}
		hash := strings.TrimSuffix(dir[i].Name(), ".mp3")
		if _, ok := permHistory[hash]; !ok {
			err := os.Remove(fmt.Sprintf("%s/%s", histFileFolder, dir[i].Name()))
			if err != nil {
				log.Println(err)
			}
		}
	}
}
