package huh

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	// "github.com/charmbracelet/huh/accessibility"
	"github.com/charmbracelet/lipgloss"
)

// ArrayList is a form field that produces multiple Input fields
type ArrayList struct {
	value *[]string
	key   string

	// customization
	title       string
	description string
	inline      bool

	// error handling
	validate func(string) error
	err      error

	// model
	textinputs []textinput.Model

	// state
	focused int

	// options
	width      int
	height     int
	accessible bool
	theme      *Theme
	keymap     InputKeyMap
}

// NewArrayList returns a new ArrayList field.
func NewArrayList() *ArrayList {
	input := textinput.New()

	i := &ArrayList{
		value:      &[]string{},
		textinputs: []textinput.Model{input},
		validate:   func(string) error { return nil },
		theme:      ThemeCharm(),
	}

	return i
}

// Value sets the value of the ArrayList field.
func (a *ArrayList) Value(value *[]string) *ArrayList {
	a.value = value
	for i := range a.textinputs {
		*a.value = append(*a.value, "")
		a.textinputs[i].SetValue("")
	}
	return a
}

// Key sets the key of the ArrayList field.
func (a *ArrayList) Key(key string) *ArrayList {
	a.key = key
	return a
}

// Title sets the title of the ArrayList field.
func (a *ArrayList) Title(title string) *ArrayList {
	a.title = title
	return a
}

// Description sets the description of the ArrayList field.
func (a *ArrayList) Description(description string) *ArrayList {
	a.description = description
	return a
}

// Prompt sets the prompt of the ArrayList field.
func (a *ArrayList) Prompt(prompt string) *ArrayList {
	for i := range a.textinputs {
		a.textinputs[i].Prompt = prompt
	}
	return a
}

// CharLimit sets the character limit of the ArrayList field.
func (a *ArrayList) CharLimit(charlimit int) *ArrayList {
	for i := range a.textinputs {
		a.textinputs[i].CharLimit = charlimit
	}
	return a
}

// Suggestions sets the suggestions to display for autocomplete in the input
// field.
func (a *ArrayList) Suggestions(suggestions []string) *ArrayList {
	for i := range a.textinputs {
		a.textinputs[i].ShowSuggestions = len(suggestions) > 1
		a.textinputs[i].KeyMap.AcceptSuggestion.SetEnabled(len(suggestions) > 0)
		a.textinputs[i].SetSuggestions(suggestions)
	}
	return a
}

// EchoMode sets the echo mode of the input.
func (a *ArrayList) EchoMode(mode EchoMode) *ArrayList {
	for i := range a.textinputs {
		a.textinputs[i].EchoMode = textinput.EchoMode(mode)
	}
	return a
}

// Password sets whether or not to hide the input while the user is typing.
//
// Deprecated: use EchoMode(EchoPassword) instead.
func (a *ArrayList) Password(password bool) *ArrayList {
	mode := textinput.EchoPassword
	if !password {
		mode = textinput.EchoNormal
	}

	for i := range a.textinputs {
		a.textinputs[i].EchoMode = mode
	}
	return a
}

// Placeholder sets the placeholder of the text input.
func (a *ArrayList) Placeholder(str string) *ArrayList {
	for i := range a.textinputs {
		a.textinputs[i].Placeholder = str
	}
	return a
}

// Inline sets whether the title and input should be on the same line.
func (a *ArrayList) Inline(inline bool) *ArrayList {
	a.inline = inline
	return a
}

// Validate sets the validation function of the ArrayList field.
func (a *ArrayList) Validate(validate func(string) error) *ArrayList {
	a.validate = validate
	return a
}

// Error returns the error of the ArrayList field.
func (a *ArrayList) Error() error {
	return a.err
}

// Skip returns whether the input should be skipped or should be blocking.
func (*ArrayList) Skip() bool {
	return false
}

// Zoom returns whether the input should be zoomed.
func (*ArrayList) Zoom() bool {
	return false
}

// Focus focuses the ArrayList field, specifically the last input field in the list
func (a *ArrayList) Focus() tea.Cmd {
	a.focused = len(a.textinputs) - 1
	return a.textinputs[a.focused].Focus()
}

// Blur blurs the ArrayList field.
func (a *ArrayList) Blur() tea.Cmd {
	a.focused = -1
	for i := range a.textinputs {
		(*a.value)[i] = a.textinputs[i].Value()
		a.textinputs[i].Blur()
		a.err = a.validate((*a.value)[i])
		if a.err != nil {
			break
		}
	}
	return nil
}

// KeyBinds returns the help message for the ArrayList field.
func (a *ArrayList) KeyBinds() []key.Binding {
	if a.textinputs[0].ShowSuggestions {
		return []key.Binding{a.keymap.AcceptSuggestion, a.keymap.Prev, a.keymap.Submit, a.keymap.Next}
	}
	return []key.Binding{a.keymap.Prev, a.keymap.Submit, a.keymap.Next}
}

// Init initializes the ArrayList field.
func (a *ArrayList) Init() tea.Cmd {
	for i := range a.textinputs {
		a.textinputs[i].Blur()
	}
	return nil
}

// Update updates the ArrayList field.
func (a *ArrayList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if a.focused == -1 {
		return a, nil
	}

	a.textinputs[a.focused], cmd = a.textinputs[a.focused].Update(msg)
	cmds = append(cmds, cmd)
	(*a.value)[a.focused] = a.textinputs[a.focused].Value()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		a.err = nil

		switch {
		case key.Matches(msg, a.keymap.Prev):
			value := a.textinputs[a.focused].Value()
			a.err = a.validate(value)
			if a.err != nil {
				return a, nil
			}
			if a.focused == 0 {
				cmds = append(cmds, prevField)
			} else {
				a.focused--
			}
		case key.Matches(msg, a.keymap.Next, a.keymap.Submit):
			value := a.textinputs[a.focused].Value()
			a.err = a.validate(value)
			if a.focused == len(a.textinputs)-1 {
				if a.err != nil && a.textinputs[a.focused].Value() != "" {
					return a, nil
				}
				if a.textinputs[a.focused].Value() == "" {
					cmds = append(cmds, nextField)
				} else {
					cmds = append(cmds, a.extendArrayList())
				}
			} else {
				if a.err != nil {
					return a, nil
				}
				a.focused++
			}
		}
	}

	return a, tea.Batch(cmds...)
}

// View renders the ArrayList field.
func (a *ArrayList) View() string {
	var sb strings.Builder

	for i := range a.textinputs {
		var ib strings.Builder
		styles := a.theme.Blurred
		if a.focused == i {
			styles = a.theme.Focused
		}

		// NB: since the method is on a pointer receiver these are being mutated.
		// Because this runs on every render this shouldn't matter in practice,
		// however.
		a.textinputs[i].PlaceholderStyle = styles.TextInput.Placeholder
		a.textinputs[i].PromptStyle = styles.TextInput.Prompt
		a.textinputs[i].Cursor.Style = styles.TextInput.Cursor
		a.textinputs[i].TextStyle = styles.TextInput.Text

		if i > 0 {
			ib.WriteString("\n")
		}

		if i == 0 {
			if a.title != "" {
				ib.WriteString(styles.Title.Render(a.title))
				if !a.inline {
					ib.WriteString("\n")
				}
			}
			if a.description != "" {
				ib.WriteString(styles.Description.Render(a.description))
				if !a.inline {
					ib.WriteString("\n")
				}
			}
		} else if a.inline {
			if a.title != "" {
				ib.WriteString(styles.Title.Render(strings.Repeat(" ", len(a.title))))
			}
			if a.description != "" {
				ib.WriteString(styles.Title.Render(strings.Repeat(" ", len(a.description))))
			}
		}

		ib.WriteString(a.textinputs[i].View())
		sb.WriteString(styles.Base.Render(ib.String()))
	}

	return sb.String()
}

// Run runs the ArrayList field in accessible mode.
func (a *ArrayList) Run() error {
	if a.accessible {
		return a.runAccessible()
	}
	return a.run()
}

// run runs the ArrayList field.
func (a *ArrayList) run() error {
	return Run(a)
}

// runAccessible runs the ArrayList field in accessible mode.
func (a *ArrayList) runAccessible() error {
	// TODO: this needs working out
	panic("ArrayList does not yet support accessability mode")
	fmt.Println(a.theme.Blurred.Base.Render(a.theme.Focused.Title.Render(a.title)))
	fmt.Println()
	// *a.value = accessibility.PromptString("Input: ", i.validate)
	// fmt.Println(i.theme.Focused.SelectedOption.Render("Input: " + *i.value + "\n"))
	return nil
}

// WithKeyMap sets the keymap on an ArrayList field.
func (a *ArrayList) WithKeyMap(k *KeyMap) Field {
	a.keymap = k.Input
	for i := range a.textinputs {
		a.textinputs[i].KeyMap.AcceptSuggestion = a.keymap.AcceptSuggestion
	}
	return a
}

// WithAccessible sets the accessible mode of the ArrayList field.
func (a *ArrayList) WithAccessible(accessible bool) Field {
	a.accessible = accessible
	return a
}

// WithTheme sets the theme of the ArrayList field.
func (a *ArrayList) WithTheme(theme *Theme) Field {
	a.theme = theme
	return a
}

// WithWidth sets the width of the ArrayList field.
func (a *ArrayList) WithWidth(width int) Field {
	a.width = width
	frameSize := a.theme.Blurred.Base.GetHorizontalFrameSize()
	titleWidth := lipgloss.Width(a.theme.Focused.Title.Render(a.title))
	descriptionWidth := lipgloss.Width(a.theme.Focused.Description.Render(a.description))
	for i := range a.textinputs {
		promptWidth := lipgloss.Width(a.textinputs[i].PromptStyle.Render(a.textinputs[i].Prompt))
		a.textinputs[i].Width = width - frameSize - promptWidth - 1
		if a.inline {
			a.textinputs[i].Width -= titleWidth
			a.textinputs[i].Width -= descriptionWidth
		}
	}
	return a
}

// WithHeight sets the height of the ArrayList field.
func (a *ArrayList) WithHeight(height int) Field {
	a.height = height
	return a
}

// WithPosition sets the position of the ArrayList field.
func (a *ArrayList) WithPosition(p FieldPosition) Field {
	a.keymap.Prev.SetEnabled(!p.IsFirst())
	a.keymap.Next.SetEnabled(!p.IsLast())
	a.keymap.Submit.SetEnabled(p.IsLast())
	return a
}

// GetKey returns the key of the field.
func (a *ArrayList) GetKey() string {
	return a.key
}

// GetValue returns the value of the field.
func (a *ArrayList) GetValue() any {
	return *a.value
}

func (a *ArrayList) extendArrayList() tea.Cmd {
	a.textinputs[a.focused].Blur()
	input := textinput.New()

	input.SetValue("")
	input.Prompt = a.textinputs[0].Prompt
	input.CharLimit = a.textinputs[0].CharLimit
	input.EchoMode = a.textinputs[0].EchoMode
	input.Placeholder = a.textinputs[0].Placeholder

	input.ShowSuggestions = a.textinputs[0].ShowSuggestions
	input.KeyMap.AcceptSuggestion.SetEnabled(a.textinputs[0].KeyMap.AcceptSuggestion.Enabled())
	input.SetSuggestions(a.textinputs[0].AvailableSuggestions())

	a.textinputs = append(a.textinputs, input)
	(*a.value) = append((*a.value), "")

	a.height++
	a.focused++
	return a.Focus()
}
