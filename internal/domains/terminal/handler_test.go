package terminal_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nootey/observv/internal/apperr"
	"github.com/nootey/observv/internal/domains/terminal"
	"github.com/nootey/observv/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type TerminalHandlerTestSuite struct {
	suite.Suite
	router      *chi.Mux
	mockService *mocks.MockTerminalServiceInterface
	handler     *terminal.TerminalHandler
}

func (suite *TerminalHandlerTestSuite) SetupTest() {
	suite.mockService = mocks.NewMockTerminalServiceInterface(suite.T())
	suite.handler = terminal.NewTerminalHandler(suite.mockService, apperr.NewResponder(zap.NewNop()))

	// Setup router
	suite.router = chi.NewRouter()
	suite.router.Mount("/terminals", suite.handler.Routes())
}

func (suite *TerminalHandlerTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
}

func TestTerminalHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(TerminalHandlerTestSuite))
}

func (suite *TerminalHandlerTestSuite) TestListTerminals_Success() {
	suite.mockService.On("ListTerminals", mock.Anything).Return([]*terminal.Terminal{
		{ID: 1, Name: "terminal-01", Status: terminal.StatusOnline},
		{ID: 2, Name: "terminal-02", Status: terminal.StatusOffline},
	}, nil)

	rec := suite.do(http.MethodGet, "/terminals/", "")

	suite.Equal(http.StatusOK, rec.Code)

	var body []terminal.Terminal
	suite.Require().NoError(json.NewDecoder(rec.Body).Decode(&body))
	suite.Len(body, 2)
	suite.Equal("terminal-01", body[0].Name)
}

func (suite *TerminalHandlerTestSuite) TestGetTerminal_InvalidID() {
	rec := suite.do(http.MethodGet, "/terminals/abc", "")

	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *TerminalHandlerTestSuite) TestGetTerminal_NotFound() {
	suite.mockService.On("GetTerminal", mock.Anything, int64(9)).
		Return(nil, apperr.New(apperr.KindNotFound, "terminal not found"))

	rec := suite.do(http.MethodGet, "/terminals/9", "")

	suite.Equal(http.StatusNotFound, rec.Code)
}

func (suite *TerminalHandlerTestSuite) TestRecordHeartbeat_Success() {
	suite.mockService.On("RecordHeartbeat", mock.Anything, int64(1), terminal.StatusOnline).
		Return(&terminal.Terminal{ID: 1, Status: terminal.StatusOnline}, nil)

	rec := suite.do(http.MethodPost, "/terminals/1/heartbeat", `{"status":"online"}`)

	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *TerminalHandlerTestSuite) TestRecordHeartbeat_InvalidBody() {
	rec := suite.do(http.MethodPost, "/terminals/1/heartbeat", `not json`)

	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *TerminalHandlerTestSuite) TestTransfer_Success() {
	suite.mockService.On("Transfer", mock.Anything, int64(1), int64(2), int64(500)).
		Return(&terminal.Terminal{ID: 1, Status: terminal.StatusOnline, DataSentBytes: 500}, nil)

	rec := suite.do(http.MethodPost, "/terminals/1/transfers", `{"to_id":2,"bytes":500}`)

	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *TerminalHandlerTestSuite) TestTransfer_PeerOffline() {
	suite.mockService.On("Transfer", mock.Anything, int64(1), int64(2), int64(500)).
		Return(nil, apperr.New(apperr.KindConflict, "terminal is offline"))

	rec := suite.do(http.MethodPost, "/terminals/1/transfers", `{"to_id":2,"bytes":500}`)

	suite.Equal(http.StatusConflict, rec.Code)
}

func (suite *TerminalHandlerTestSuite) do(method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	suite.router.ServeHTTP(rec, req)
	return rec
}
