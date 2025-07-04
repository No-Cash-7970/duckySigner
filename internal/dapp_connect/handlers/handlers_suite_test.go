package handlers_test

import (
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/gommon/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/wailsapp/wails/v3/pkg/application"

	"duckysigner/internal/testing/mocks"
	. "duckysigner/services"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DApp Connect Endpoints Suite")
}

func getResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

var dcService DappConnectService

const defaultUserRespTimeout = 2 * time.Second
const rootGetPort = "1384"
const sessionInitPostPort = "1385"

func setUpDcService(port string, mockSessionKey string) {
	dcService = DappConnectService{
		// Make sure to use a port that is not used in another test so the
		// tests can be run in parallel
		ServerAddr:       ":" + port,
		LogLevel:         log.ERROR,
		HideServerBanner: true,
		HideServerPort:   true,
		WailsApp: application.New(application.Options{
			LogLevel: slog.LevelError,
		}),
		UserResponseTimeout: defaultUserRespTimeout,
		ECDHCurve:           &mocks.EcdhCurveMock{GeneratedPrivateKey: mockSessionKey},
	}
	By("Starting dApp connect server")
	dcService.Start()
	DeferCleanup(func() {
		By("Stopping dApp connect server")
		dcService.Stop()
	})
}
