package e2e

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/RevoTale/no-js/internal/filesystem"
	"github.com/stretchr/testify/require"
)

type responseSnapshot struct {
	Status               int
	Body                 string
	ContentType          string
	HXTriggerAfterSettle string
	Location             string
	Allow                string
	XMainMiddleware      string
}

type requestOptions struct {
	Headers map[string]string
	Host    string
}

type fixtureServer struct {
	AppDir  string
	BaseURL string
	client  *http.Client
	cmd     *exec.Cmd
	waitCh  chan error
	stdout  bytes.Buffer
	stderr  bytes.Buffer
}

var (
	noJSToolOnce sync.Once
	noJSToolPath string
	noJSToolErr  error
)

func existingTemplgenPaths(t *testing.T, appDir string, paths ...string) []string {
	t.Helper()

	out := make([]string, 0, len(paths))
	for _, path := range paths {
		info, err := os.Stat(filepath.Join(appDir, filepath.FromSlash(path)))
		if os.IsNotExist(err) {
			continue
		}
		require.NoError(t, err)
		require.True(t, info.IsDir(), "%s must be a directory", path)
		out = append(out, path)
	}

	return out
}

func prepareFixtureApp(t *testing.T, fixtureName string) string {
	t.Helper()

	appDir := copyFixtureApp(t, fixtureName)

	runGo(t, appDir, "tool", "no-js", "gen", "routes", "-root", ".")
	templgenArgs := []string{
		"tool",
		"templgen",
		"-base",
		".",
	}
	for _, path := range existingTemplgenPaths(t, appDir, "web/generated", "web/components", "web/view") {
		templgenArgs = append(templgenArgs, "-path", path)
	}
	runGo(t, appDir, templgenArgs...)
	runGo(t, appDir, "tool", "no-js", "gen", "assets", "-root", ".")

	return appDir
}

func prepareFixtureAppWithNoJSGen(t *testing.T, fixtureName string) string {
	t.Helper()

	return prepareFixtureAppWithNoJSGenConfig(t, fixtureName, "")
}

func prepareFixtureAppWithNoJSGenConfig(t *testing.T, fixtureName string, bundleConfig string) string {
	t.Helper()

	appDir := copyFixtureApp(t, fixtureName)
	if strings.TrimSpace(bundleConfig) != "" {
		require.NoError(t, os.WriteFile(filepath.Join(appDir, "no-js.bundle.yaml"), []byte(bundleConfig), 0o644))
	}
	runGo(t, appDir, "tool", "no-js", "gen", "-root", ".")

	return appDir
}

func copyFixtureApp(t *testing.T, fixtureName string) string {
	t.Helper()

	repoRoot := repoRootPath(t)
	appDir := filepath.Join(t.TempDir(), fixtureName)

	require.NoError(
		t,
		filesystem.CopyTree(filepath.Join(repoRoot, "e2e", "testdata", fixtureName), appDir),
	)
	writeFixtureModuleFiles(t, repoRoot, appDir)

	return appDir
}

func startPreparedFixture(t *testing.T, fixtureName string) (string, *fixtureServer) {
	t.Helper()

	appDir := prepareFixtureApp(t, fixtureName)
	return appDir, startBuiltFixtureServer(t, appDir)
}

func startPreparedFixtureWithNoJSGen(t *testing.T, fixtureName string) (string, *fixtureServer) {
	t.Helper()

	appDir := prepareFixtureAppWithNoJSGen(t, fixtureName)
	return appDir, startBuiltFixtureServer(t, appDir)
}

func startPreparedFixtureWithNoJSGenConfig(
	t *testing.T,
	fixtureName string,
	bundleConfig string,
) (string, *fixtureServer) {
	t.Helper()

	appDir := prepareFixtureAppWithNoJSGenConfig(t, fixtureName, bundleConfig)
	return appDir, startBuiltFixtureServer(t, appDir)
}

func startBuiltFixtureServer(t *testing.T, appDir string) *fixtureServer {
	t.Helper()

	serverBin := filepath.Join(t.TempDir(), "fixture-server")
	runGo(t, appDir, "build", "-o", serverBin, "./cmd/server")

	cmd := exec.Command(serverBin, "-addr", "127.0.0.1:0")
	cmd.Dir = appDir
	cmd.Env = append(os.Environ(), "GOWORK=off")

	stdoutPipe, err := cmd.StdoutPipe()
	require.NoError(t, err)

	stderrPipe, err := cmd.StderrPipe()
	require.NoError(t, err)

	server := &fixtureServer{
		AppDir: appDir,
		client: &http.Client{
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{DisableCompression: true},
		},
		cmd:    cmd,
		waitCh: make(chan error, 1),
	}

	listenCh := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			server.stdout.WriteString(line)
			server.stdout.WriteByte('\n')
			if strings.HasPrefix(line, "LISTEN_URL=") {
				select {
				case listenCh <- strings.TrimSpace(strings.TrimPrefix(line, "LISTEN_URL=")):
				default:
				}
			}
		}
	}()

	go func() {
		_, _ = io.Copy(&server.stderr, stderrPipe)
	}()

	require.NoError(t, cmd.Start())
	go func() {
		server.waitCh <- cmd.Wait()
	}()

	select {
	case baseURL := <-listenCh:
		server.BaseURL = strings.TrimRight(baseURL, "/")
	case err := <-server.waitCh:
		require.NoError(t, err, "fixture server exited early\n%s", server.failureContext())
	case <-time.After(10 * time.Second):
		t.Fatalf("timed out waiting for fixture server readiness\n%s", server.failureContext())
	}

	t.Cleanup(func() {
		stopFixtureServer(t, server)
	})

	return server
}

func stopFixtureServer(t *testing.T, server *fixtureServer) {
	t.Helper()

	if server == nil || server.cmd == nil || server.cmd.Process == nil {
		return
	}

	select {
	case err := <-server.waitCh:
		if err != nil {
			require.NoError(t, err, "fixture server exited with error\n%s", server.failureContext())
		}
		return
	default:
	}

	_ = server.cmd.Process.Signal(os.Interrupt)

	select {
	case err := <-server.waitCh:
		if err != nil {
			require.NoError(t, err, "fixture server exited with error\n%s", server.failureContext())
		}
	case <-time.After(5 * time.Second):
		_ = server.cmd.Process.Kill()
		select {
		case err := <-server.waitCh:
			if err != nil && !strings.Contains(err.Error(), "killed") {
				require.NoError(t, err, "fixture server kill failed\n%s", server.failureContext())
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for fixture server shutdown\n%s", server.failureContext())
		}
	}
}

func requestFixture(
	t *testing.T,
	server *fixtureServer,
	method string,
	requestPath string,
	body io.Reader,
	opts requestOptions,
) responseSnapshot {
	t.Helper()

	req, err := http.NewRequest(method, requestURL(server.BaseURL, requestPath), body)
	require.NoError(t, err)

	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}
	if opts.Host != "" {
		req.Host = opts.Host
	}

	resp, err := server.client.Do(req)
	require.NoError(t, err, "request failed for %s %s\n%s", method, requestPath, server.failureContext())
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return responseSnapshot{
		Status:               resp.StatusCode,
		Body:                 string(responseBody),
		ContentType:          resp.Header.Get("Content-Type"),
		HXTriggerAfterSettle: resp.Header.Get("HX-Trigger-After-Settle"),
		Location:             resp.Header.Get("Location"),
		Allow:                resp.Header.Get("Allow"),
		XMainMiddleware:      resp.Header.Get("X-Main-Middleware"),
	}
}

func requestStreamFixture(
	t *testing.T,
	server *fixtureServer,
	requestPath string,
	opts requestOptions,
	releasePath string,
) streamSnapshot {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, requestURL(server.BaseURL, requestPath), nil)
	require.NoError(t, err)

	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}
	if opts.Host != "" {
		req.Host = opts.Host
	}

	resp, err := server.client.Do(req)
	require.NoError(t, err, "stream request failed for %s\n%s", requestPath, server.failureContext())
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	type firstChunkResult struct {
		data []byte
		err  error
	}

	firstChunkCh := make(chan firstChunkResult, 1)
	go func() {
		buf := make([]byte, 64)
		n, readErr := resp.Body.Read(buf)
		firstChunkCh <- firstChunkResult{data: buf[:n], err: readErr}
	}()

	var firstChunk []byte
	select {
	case result := <-firstChunkCh:
		require.True(t, result.err == nil || result.err == io.EOF, "unexpected stream read error: %v", result.err)
		require.NotEmpty(t, result.data)
		firstChunk = result.data
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for first stream chunk\n%s", server.failureContext())
	}

	release := requestFixture(t, server, http.MethodPost, releasePath, nil, requestOptions{})
	require.Equal(t, http.StatusNoContent, release.Status)

	rest, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	bodyText := string(firstChunk) + string(rest)
	return streamSnapshot{
		Status:      resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		FirstChunk:  string(firstChunk),
		Body:        bodyText,
	}
}

func requestURL(baseURL string, requestPath string) string {
	if strings.HasPrefix(requestPath, "http://") || strings.HasPrefix(requestPath, "https://") {
		return requestPath
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(requestPath, "/")
}

func (server *fixtureServer) failureContext() string {
	return "stdout:\n" + server.stdout.String() + "\nstderr:\n" + server.stderr.String()
}

func writeFixtureModuleFiles(t *testing.T, repoRoot string, appDir string) {
	t.Helper()

	goModPath := filepath.Join(appDir, "go.mod")
	goMod, err := os.ReadFile(goModPath)
	require.NoError(t, err)

	replacePattern := regexp.MustCompile(`(?m)^replace github.com/RevoTale/no-js => .+$`)
	rewrittenGoMod := replacePattern.ReplaceAllString(
		string(goMod),
		"replace github.com/RevoTale/no-js => "+filepath.ToSlash(repoRoot),
	)
	require.NotEqual(t, string(goMod), rewrittenGoMod, "fixture go.mod must declare no-js replace")

	require.NoError(t, os.WriteFile(goModPath, []byte(rewrittenGoMod), 0o644))
}

func runGo(t *testing.T, dir string, args ...string) []byte {
	t.Helper()

	cmd := exec.Command("go", args...)
	cmd.Dir = dir

	goCacheDir := filepath.Join(dir, ".cache", "go-build")
	require.NoError(t, os.MkdirAll(goCacheDir, 0o755))

	cmd.Env = append(
		os.Environ(),
		"GOWORK=off",
		"GOCACHE="+goCacheDir,
	)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", strings.TrimSpace(string(output)))
	return output
}

func runNoJSError(t *testing.T, dir string, args ...string) []byte {
	t.Helper()

	cmd := exec.Command(noJSToolBinary(t), args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.Error(t, err, "command unexpectedly succeeded: no-js %s", strings.Join(args, " "))
	return output
}

func noJSToolBinary(t *testing.T) string {
	t.Helper()

	noJSToolOnce.Do(func() {
		repoRoot := repoRootPath(t)
		outDir, err := os.MkdirTemp("", "no-js-e2e-tool-*")
		if err != nil {
			noJSToolErr = fmt.Errorf("create no-js tool temp dir: %w", err)
			return
		}
		goCacheDir := filepath.Join(outDir, "go-build")
		if err := os.MkdirAll(goCacheDir, 0o755); err != nil {
			noJSToolErr = fmt.Errorf("create no-js tool build cache: %w", err)
			return
		}
		noJSToolPath = filepath.Join(outDir, "no-js")
		cmd := exec.Command("go", "build", "-o", noJSToolPath, "./cmd/no-js")
		cmd.Dir = repoRoot
		cmd.Env = append(
			os.Environ(),
			"GOWORK=off",
			"GOCACHE="+goCacheDir,
		)
		output, buildErr := cmd.CombinedOutput()
		if buildErr != nil {
			noJSToolErr = fmt.Errorf("build no-js tool: %w: %s", buildErr, strings.TrimSpace(string(output)))
		}
	})
	require.NoError(t, noJSToolErr)
	return noJSToolPath
}

func repoRootPath(t *testing.T) string {
	t.Helper()

	_, fileName, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(fileName), ".."))
}
