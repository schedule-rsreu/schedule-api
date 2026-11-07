//nolint:testpackage // Exercise the package-private middleware after the Echo upgrade.
package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestRequestLoggingPreservesStatusAndRedactsParams(t *testing.T) {
	for _, test := range []struct {
		method, path string
		status       int
	}{
		{http.MethodGet, "/missing", http.StatusNotFound},
		{http.MethodPost, "/groups/344", http.StatusMethodNotAllowed},
		{http.MethodGet, "/groups/344?token=private", http.StatusForbidden},
	} {
		t.Run(test.method+test.path, func(t *testing.T) {
			var output bytes.Buffer
			logger := zerolog.New(&output)
			e := echo.New()
			setupLogger(e, &logger)
			e.GET("/groups/:group", func(c *echo.Context) error { return echo.ErrForbidden })
			recorder := httptest.NewRecorder()
			e.ServeHTTP(recorder, httptest.NewRequest(test.method, test.path, http.NoBody))
			require.Equal(t, test.status, recorder.Code)
			var entry struct {
				Params map[string][]string `json:"params"`
				Status int                 `json:"status"`
			}
			require.NoError(t, json.Unmarshal(output.Bytes(), &entry))
			require.Equal(t, test.status, entry.Status)
			if test.status == http.StatusForbidden {
				require.Equal(t, []string{"344"}, entry.Params["group"])
				require.Equal(t, []string{"[REDACTED]"}, entry.Params["token"])
			}
		})
	}
}
