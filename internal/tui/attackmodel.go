package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dilyxs/CannonCLI/pkg"
)

type AttackModel struct {
	RequestPerSecond    int
	HowManySecond       int
	ReqDetails          pkg.RequestDetails
	WorkersCount        int
	Cancel              chan bool
	ResponseChan        chan pkg.ResponseWithStatus
	Results             []pkg.ResponseWithStatus
	FailedResponses     int
	SuccessfulResponses int
	EnOfListener        chan pkg.EndOfSequence
}
type SequnceFinished string

func (m AttackModel) ListenForIncomingMessages() tea.Cmd {
	return func() tea.Msg {
		res, ok := <-m.ResponseChan
		if ok {
			return res
		}
		return nil
	}
}

func (m AttackModel) ListenForEnding() tea.Cmd {
	return func() tea.Msg {
		res, ok := <-m.EnOfListener
		if ok {
			return res
		}
		return nil
	}
}

func InitModel(RequestPerSecond, HowManySecond, WorkersCount int, ReqDetails pkg.RequestDetails, CancelChan chan bool, OutputChan chan pkg.ResponseWithStatus) AttackModel {
	return AttackModel{
		RequestPerSecond: RequestPerSecond,
		HowManySecond:    HowManySecond,
		ReqDetails:       ReqDetails,
		WorkersCount:     WorkersCount,
		Cancel:           CancelChan,
		ResponseChan:     OutputChan,
	}
}

func (m AttackModel) Init() tea.Cmd {
	return m.ListenForIncomingMessages()
}

func (m AttackModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case pkg.ResponseWithStatus:
		if msg.Error != nil && msg.IsOk {
			m.SuccessfulResponses += 1
		} else {
			m.FailedResponses += 1
		}
		m.Results = append(m.Results, msg)
		return nil, m.ListenForIncomingMessages()
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
	return fmt.Sprintf("current have %d succes and %d faillure", m.SuccessfulResponses, m.FailedResponses)
}
