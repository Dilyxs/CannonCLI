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

func InitModel(RequestPerSecond, HowManySecond, WorkersCount int, ReqDetails pkg.RequestDetails, CancelChan chan bool, OutputChan chan pkg.ResponseWithStatus) AttackModel {
	return AttackModel{
		RequestPerSecond:    RequestPerSecond,
		HowManySecond:       HowManySecond,
		ReqDetails:          ReqDetails,
		WorkersCount:        WorkersCount,
		Cancel:              CancelChan,
		ResponseChan:        OutputChan,
		Results:             make([]pkg.ResponseWithStatus, 0),
		FailedResponses:     0,
		SuccessfulResponses: 0,
		EnOfListener:        make(chan pkg.EndOfSequence, 2),
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
		if msg.Error == nil && msg.IsOk {
			m.SuccessfulResponses += 1
		} else {
			m.FailedResponses += 1
		}
		m.Results = append(m.Results, msg)
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
	return fmt.Sprintf("current have %d succes and %d faillure", m.SuccessfulResponses, m.FailedResponses)
}
