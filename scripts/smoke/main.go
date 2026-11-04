// Cross-platform CLI smoke: run against a built procpane binary (CI + local).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/zx06/procpane/internal/model"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("CLI smoke OK")
}

func run() error {
	bin := procpaneBin()
	if _, err := os.Stat(bin); err != nil {
		return fmt.Errorf("procpane binary %q: %w (build first)", bin, err)
	}

	fmt.Println("== list --json ==")
	out, err := runProcpane(bin, "list", "--json")
	if err != nil {
		return fmt.Errorf("list --json: %w", err)
	}
	var list model.ListResponse
	if err := json.Unmarshal(out, &list); err != nil {
		return fmt.Errorf("parse list json: %w", err)
	}
	if list.Listeners == nil {
		return fmt.Errorf("list json: listeners field missing")
	}

	fmt.Println("== kill 1 refused ==")
	if err := runProcpaneExpectFail(bin, "kill", "1"); err != nil {
		return err
	}

	port := testPort()
	ln, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return fmt.Errorf("listen :%d: %w", port, err)
	}
	defer ln.Close()

	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})}
	go func() { _ = srv.Serve(ln) }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	time.Sleep(400 * time.Millisecond)

	portSpec := fmt.Sprintf(":%d", port)
	fmt.Printf("== find %s ==\n", portSpec)
	findOut, err := runProcpane(bin, "find", portSpec)
	if err != nil {
		return fmt.Errorf("find: %w", err)
	}
	if !strings.Contains(string(findOut), portSpec) {
		return fmt.Errorf("find output does not mention %s:\n%s", portSpec, findOut)
	}

	pid := os.Getpid()
	fmt.Printf("== show %d ==\n", pid)
	showOut, err := runProcpane(bin, "show", strconv.Itoa(pid))
	if err != nil {
		return fmt.Errorf("show: %w", err)
	}
	var info model.ProcessInfo
	if err := json.Unmarshal(showOut, &info); err != nil {
		return fmt.Errorf("parse show json: %w", err)
	}
	if info.PID != int32(pid) {
		return fmt.Errorf("show pid: got %d want %d", info.PID, pid)
	}
	return nil
}

func procpaneBin() string {
	if v := os.Getenv("PROCPANE_BIN"); v != "" {
		return v
	}
	name := "procpane"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join("bin", name)
}

func testPort() int {
	if v := os.Getenv("PROCPANE_TEST_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p <= 0 || p > 65535 {
			fmt.Fprintf(os.Stderr, "invalid PROCPANE_TEST_PORT %q, using default\n", v)
		} else {
			return p
		}
	}
	return 18765
}

func runProcpane(bin string, args ...string) ([]byte, error) {
	cmd := exec.Command(bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %v: %w\nstderr: %s", filepath.Base(bin), args, err, stderr.Bytes())
	}
	return stdout.Bytes(), nil
}

func runProcpaneExpectFail(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err == nil {
		return fmt.Errorf("%s %v: expected non-zero exit", filepath.Base(bin), args)
	}
	return nil
}
