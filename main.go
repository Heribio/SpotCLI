package main

import (
    "fmt"
    "os"
    "net/http"
    "io"
    "encoding/json"
    "os/exec"
    tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/termenv"
    "time"
)

func main() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Alas, there's been an error: %v", err)
        os.Exit(1)
    }
}

func getSong(link string) {
    resp, err := http.Get("https://pipedapi.syncpundit.io/streams/" + link)
    if err != nil {
        fmt.Printf("Alas, there's been an error: %v", err)
        os.Exit(1)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("Alas, there's been an error reading the body: %v", err)
        os.Exit(1)
    }

    type AudioStream struct {
		URL       string `json:"url"`
		Format    string `json:"format"`
		Quality   string `json:"quality"`
		MimeType  string `json:"mimeType"`
		Codec     string `json:"codec"`
		Bitrate   int    `json:"bitrate"`
		ContentLength int `json:"contentLength"`
	}

    type Result struct {
        Title string `json:"title"`
        Artist string `json:"uploader"`
        Duration int `json:"duration"`
        AudioLink []AudioStream `json:"audioStreams"`
    }
    
    var result Result
    json.Unmarshal(body, &result)

    if len(result.AudioLink) <= 0 {
        fmt.Println("No audio stream found")
        os.Exit(1)
    }

    fmt.Println(result.Title)
    fmt.Println(result.Artist)
    fmt.Println(result.AudioLink[0].URL)
    fmt.Println("Duration: ", result.Duration)
    downloadSong(result.AudioLink[1].URL, "audio.mp4", result.Duration)
}

func downloadSong(url, filename string, duration int) {
    fmt.Println("Downloading audio...")
    cmd := exec.Command("curl", "-o", filename, url)
    err := cmd.Run()
    if err != nil {
        fmt.Printf("Error downloading audio stream: %v\n", err)
        return
    }

    fmt.Printf("Audio saved as %s\n", filename)
    playCmd := exec.Command("afplay", filename, "-t", fmt.Sprintf("%d", duration))
    err = playCmd.Run()
    if err != nil {
        fmt.Printf("Error playing audio file: %v\n", err)
        return
    }
    time.Sleep(time.Duration(duration) * time.Second)
    stopPlay := exec.Command("killall", "afplay")
    err = stopPlay.Run()
    if err != nil {
        fmt.Printf("Error stopping audio playback: %v\n", err)
        return
    }
    os.Remove(filename)

    fmt.Println("Audio playback complete")
}


var (
	subtle        = makeFgStyle("241")
	term          = termenv.EnvColorProfile()
)

func initialModel() menu {
    return menu{
        choices: []string{"quick", "playlists", "settings", "help", "exit"},
        selected: map[int]struct{}{},
    }
}

type menu struct {
    choices  []string
    cursor   int
    selected map[int]struct{}
    Chosen   bool
    Quitting bool
}

func (m menu) Init() tea.Cmd {
    return nil
}

    
func (m menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
 	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.Quitting = true
			return m, tea.Quit
		}
	}

	if !m.Chosen {
		return updateChoices(msg, m)
	}
	return updateChosen(msg, m)
}

func updateChoices(msg tea.Msg, m menu) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
        case tea.KeyMsg:
    switch msg.String() {

        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }

        case "down", "j":
            if m.cursor < len(m.choices)-1 {
                m.cursor++
            }

        case "enter":
            m.selected[m.cursor] = struct{}{}
            m.Chosen = true   
            return m, tea.Quit
        }
    }
    return m, nil
}

func updateChosen(msg tea.Msg, m menu) (tea.Model, tea.Cmd) {
    return m, nil
}

func (m menu) View() string {
    var s string
    if m.Quitting {
        return "\n  See you later!\n\n" 
    }
    if m.Chosen {
        s = chosenView(m)
    } else {
        s = choicesView(m)
    }
	return indent.String("\n"+s+"\n\n", 2)
}

func choicesView(m menu) string {
    c := m.cursor

    choices := fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s",
		checkbox("Quick", c == 0),
		checkbox("Playlists", c == 1),
		checkbox("Settings", c == 2),
		checkbox("Help", c == 3),
        checkbox("Exit", c == 4),
	)
	return fmt.Sprintf(choices)
}

func chosenView(m menu) string {
    var msg string
    switch m.cursor {
    case 0:
        msg = fmt.Sprintf("Quick")
        getSong("qAmulKjcHoo")
    case 1:
        msg = fmt.Sprintf("Playlists")
    case 2:
        msg = fmt.Sprintf("Settings")
    case 3:
        msg = fmt.Sprintf("Help")
    case 4:
        msg = fmt.Sprintf("Exit")
        os.Exit(0)
    }
    return msg
}

func checkbox(label string, checked bool) string {
	if checked {
		return colorFg("[x] "+label, "39")
	}
	return fmt.Sprintf("[ ] %s", label)
}

func colorFg(val, color string) string {
	return termenv.String(val).Foreground(term.Color(color)).String()
}

// Return a function that will colorize the foreground of a given string.
func makeFgStyle(color string) func(string) string {
	return termenv.Style{}.Foreground(term.Color(color)).Styled
}

