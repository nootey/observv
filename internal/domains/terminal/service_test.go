package terminal_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/nootey/observv/internal/apperr"
	"github.com/nootey/observv/internal/domains/terminal"
	"github.com/nootey/observv/tests/integration"

	"github.com/stretchr/testify/suite"
)

type TerminalServiceTestSuite struct {
	integration.ServiceIntegrationSuite
}

func TestTerminalServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TerminalServiceTestSuite))
}

func (s *TerminalServiceTestSuite) TestListTerminals() {
	svc := s.TC.Services.TerminalService

	s.seedTerminals(3)

	result, err := svc.ListTerminals(context.Background())

	s.Require().NoError(err)
	s.Assert().Len(result, 3)
	s.Assert().Equal("terminal-01", result[0].Name)
	s.Assert().Equal(terminal.StatusOffline, result[0].Status)
}

func (s *TerminalServiceTestSuite) TestGetTerminal_NotFound() {
	svc := s.TC.Services.TerminalService

	_, err := svc.GetTerminal(context.Background(), 999)

	s.assertStatus(err, http.StatusNotFound)
}

func (s *TerminalServiceTestSuite) TestRecordHeartbeat_SetsStatus() {
	svc := s.TC.Services.TerminalService
	ids := s.seedTerminals(1)

	result, err := svc.RecordHeartbeat(context.Background(), ids[0], terminal.StatusOnline)

	s.Require().NoError(err)
	s.Assert().Equal(terminal.StatusOnline, result.Status)
	s.Assert().NotNil(result.LastSeenAt)
	s.Assert().Zero(result.UptimeSeconds)
}

func (s *TerminalServiceTestSuite) TestRecordHeartbeat_AddsUptimeWhileOnline() {
	svc := s.TC.Services.TerminalService
	ids := s.seedTerminals(1)
	s.setLastSeen(ids[0], terminal.StatusOnline, 30)

	result, err := svc.RecordHeartbeat(context.Background(), ids[0], terminal.StatusOnline)

	s.Require().NoError(err)
	s.Assert().InDelta(30, result.UptimeSeconds, 2)
}

func (s *TerminalServiceTestSuite) TestRecordHeartbeat_NoUptimeWhileOffline() {
	svc := s.TC.Services.TerminalService
	ids := s.seedTerminals(1)
	s.setLastSeen(ids[0], terminal.StatusOffline, 30)

	result, err := svc.RecordHeartbeat(context.Background(), ids[0], terminal.StatusOnline)

	s.Require().NoError(err)
	s.Assert().Zero(result.UptimeSeconds)
}

func (s *TerminalServiceTestSuite) TestRecordHeartbeat_InvalidStatus() {
	svc := s.TC.Services.TerminalService
	ids := s.seedTerminals(1)

	_, err := svc.RecordHeartbeat(context.Background(), ids[0], "sleeping")

	s.assertStatus(err, http.StatusBadRequest)
}

func (s *TerminalServiceTestSuite) TestRecordHeartbeat_NotFound() {
	svc := s.TC.Services.TerminalService

	_, err := svc.RecordHeartbeat(context.Background(), 999, terminal.StatusOnline)

	s.assertStatus(err, http.StatusNotFound)
}

func (s *TerminalServiceTestSuite) TestTransfer_Success() {
	svc := s.TC.Services.TerminalService
	ids := s.seedTerminals(2)
	s.setOnline(ids...)

	result, err := svc.Transfer(context.Background(), ids[0], ids[1], 500)

	s.Require().NoError(err)
	s.Assert().Equal(ids[0], result.ID)
	s.Assert().Equal(int64(500), result.DataSentBytes)
}

func (s *TerminalServiceTestSuite) TestTransfer_PeerOffline() {
	svc := s.TC.Services.TerminalService
	ids := s.seedTerminals(2)
	s.setOnline(ids[0])

	_, err := svc.Transfer(context.Background(), ids[0], ids[1], 500)

	s.assertStatus(err, http.StatusConflict)

	sender, err := svc.GetTerminal(context.Background(), ids[0])
	s.Require().NoError(err)
	s.Assert().Zero(sender.DataSentBytes)
}

func (s *TerminalServiceTestSuite) TestTransfer_ToSelf() {
	svc := s.TC.Services.TerminalService
	ids := s.seedTerminals(1)

	_, err := svc.Transfer(context.Background(), ids[0], ids[0], 500)

	s.assertStatus(err, http.StatusBadRequest)
}

func (s *TerminalServiceTestSuite) TestTransfer_InvalidBytes() {
	svc := s.TC.Services.TerminalService
	ids := s.seedTerminals(2)

	_, err := svc.Transfer(context.Background(), ids[0], ids[1], 0)

	s.assertStatus(err, http.StatusBadRequest)
}

func (s *TerminalServiceTestSuite) seedTerminals(count int) []int64 {
	for i := 1; i <= count; i++ {
		err := s.TC.Queries.CreateTerminal(context.Background(), fmt.Sprintf("terminal-%02d", i))
		s.Require().NoError(err)
	}

	terminals, err := s.TC.Queries.ListTerminals(context.Background())
	s.Require().NoError(err)

	ids := make([]int64, len(terminals))
	for i, t := range terminals {
		ids[i] = t.ID
	}
	return ids
}

func (s *TerminalServiceTestSuite) setOnline(ids ...int64) {
	for _, id := range ids {
		_, err := s.TC.Services.TerminalService.RecordHeartbeat(context.Background(), id, terminal.StatusOnline)
		s.Require().NoError(err)
	}
}

func (s *TerminalServiceTestSuite) setLastSeen(id int64, status string, secondsAgo int) {
	_, err := s.TC.DB.Exec(
		`UPDATE terminals SET status = $1, last_seen_at = NOW() - make_interval(secs => $2) WHERE id = $3`,
		status, secondsAgo, id,
	)
	s.Require().NoError(err)
}

func (s *TerminalServiceTestSuite) assertStatus(err error, want int) {
	s.Require().Error(err)
	status, _ := apperr.Resolve(err)
	s.Assert().Equal(want, status)
}
