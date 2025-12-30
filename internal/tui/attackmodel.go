package tui

import (
	"fmt"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dilyxs/CannonCLI/pkg"
)

type AttackModel struct {
	RequestsPerStatusCode map[int][]pkg.ResponseWithStatus
	RequestPerSecond      int
	HowManySecond         int
	ReqDetails            pkg.RequestDetails
	WorkersCount          int
	Cancel                chan bool
	ResponseChan          chan pkg.ResponseWithStatus
	Results               []pkg.ResponseWithStatus
	FailedResponses       int
	SuccessfulResponses   int
	EnOfListener          chan pkg.EndOfSequence
}

var (
	style2xx = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")) // Green
	style3xx = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")) // Yellow
	style4xx = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")) // Orange
	style5xx = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")) // Red

	barChar = "█"
)

const EndString pkg.EndOfSequence = "End"

func (m AttackModel) ListenForIncomingMessages() tea.Cmd {
	return func() tea.Msg {
		res, ok := <-m.ResponseChan
		if ok {
			return res
		}
		return EndString
	}
}

func (m AttackModel) ListenForEnding() tea.Cmd {
	return func() tea.Msg {
		res, ok := <-m.EnOfListener
		if ok {
			return res
		}
		return EndString
	}
}

const UnknownErrorEncountered int = 666

func InitModel(RequestPerSecond, HowManySecond, WorkersCount int, ReqDetails pkg.RequestDetails, CancelChan chan bool, OutputChan chan pkg.ResponseWithStatus) AttackModel {
	return AttackModel{
		RequestsPerStatusCode: make(map[int][]pkg.ResponseWithStatus),
		RequestPerSecond:      RequestPerSecond,
		HowManySecond:         HowManySecond,
		ReqDetails:            ReqDetails,
		WorkersCount:          WorkersCount,
		Cancel:                CancelChan,
		ResponseChan:          OutputChan,
		Results:               make([]pkg.ResponseWithStatus, 0),
		FailedResponses:       0,
		SuccessfulResponses:   0,
		EnOfListener:          make(chan pkg.EndOfSequence, 2),
	}
}

func (m AttackModel) StartAttack() {
	pkg.RunAction(m.RequestPerSecond, m.HowManySecond, m.ReqDetails, m.WorkersCount, m.Cancel, m.ResponseChan, m.EnOfListener)
}

func (m AttackModel) Init() tea.Cmd {
	go m.StartAttack()
	return tea.Batch(
		m.ListenForIncomingMessages(),
		m.ListenForEnding(),
	)
}

func (m AttackModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case pkg.ResponseWithStatus:
		m.HandleResponse(msg) // not sure if ok, since we pass in a pointer(auto does it, but should copy be returned instead?
		return m, m.ListenForIncomingMessages()
	case pkg.EndOfSequence:
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Cancel <- true
			return m, nil
		}
	}
	return m, nil
}

func (m AttackModel) View() string {
	var keys []int
	maxVal := 0

	counts := make(map[int]int)

	for k, v := range m.RequestsPerStatusCode {
		keys = append(keys, k)
		count := len(v)
		counts[k] = count
		if count > maxVal {
			maxVal = count
		}
	}
	sort.Ints(keys)

	graph := ""
	maxBarWidth := 40

	for _, code := range keys {
		count := counts[code]

		barWidth := 0
		if maxVal > 0 {
			barWidth = int((float64(count) / float64(maxVal)) * float64(maxBarWidth))
		}

		var style lipgloss.Style
		switch {
		case code >= 200 && code < 300:
			style = style2xx
		case code >= 300 && code < 400:
			style = style3xx
		case code >= 400 && code < 500:
			style = style4xx
		case code >= 500:
			style = style5xx
		default:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		}

		bar := style.Render(repeat(barChar, barWidth))
		label := style.Render(fmt.Sprintf("%3d", code))

		graph += fmt.Sprintf("%s │ %s %d\n", label, bar, count)
	}

	summary := fmt.Sprintf(
		"Total: %d | Success: %s | Failure: %s",
		m.SuccessfulResponses+m.FailedResponses,
		style2xx.Render(fmt.Sprintf("%d", m.SuccessfulResponses)),
		style5xx.Render(fmt.Sprintf("%d", m.FailedResponses)),
	)

	return fmt.Sprintf("\n%s\n\n%s\n\nPress q to quit.\n", summary, graph)
}

func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	out := ""
	for range n {
		out += s
	}
	return out
}

func (m *AttackModel) HandleResponse(msg pkg.ResponseWithStatus) {
	defer func() {
		m.Results = append(m.Results, msg)
	}()
	if msg.Error == nil && msg.IsOk {
		m.SuccessfulResponses += 1
		m.RequestsPerStatusCode[msg.R.StatusCode] = append(m.RequestsPerStatusCode[msg.R.StatusCode], msg)

	}
	// got a response back, but it's a 400 or 500 series
	if (msg.Error == nil) && (msg.R != nil) {
		m.FailedResponses += 1
		m.RequestsPerStatusCode[msg.R.StatusCode] = append(m.RequestsPerStatusCode[msg.R.StatusCode], msg)
	} else {
		m.FailedResponses += 1
		switch ErrorType := msg.Error.(type) {
		case pkg.CustomeErrors:
			m.RequestsPerStatusCode[ErrorType.ErrorCode] = append(m.RequestsPerStatusCode[ErrorType.ErrorCode], msg)
		default:
			m.RequestsPerStatusCode[UnknownErrorEncountered] = append(m.RequestsPerStatusCode[UnknownErrorEncountered], msg)
		}
	}
}
