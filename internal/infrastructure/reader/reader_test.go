package reader_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/4domm/ngxstat/internal/domain"
	"github.com/4domm/ngxstat/internal/infrastructure/reader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindFilesByPattern_SingleFile(t *testing.T) {
	fr := &reader.FileReader{}

	tmpFile, err := os.CreateTemp("", "testfile_*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	data, err := fr.FindFilesByPattern(tmpFile.Name())

	assert.NoError(t, err)
	assert.Len(t, data, 1)
	assert.NotNil(t, data[0])
}

func TestFindFilesByPattern_NoMatch(t *testing.T) {
	fr := &reader.FileReader{}

	data, err := fr.FindFilesByPattern("nonexistent_file.txt")

	assert.Nil(t, data)
	assert.ErrorIs(t, err, domain.ErrFinding)
}

func TestFindFilesByPattern(t *testing.T) {
	read := &reader.FileReader{}

	t.Run("Files Found", func(t *testing.T) {
		tmpDir := t.TempDir()

		file1 := filepath.Join(tmpDir, "file123.txt")
		file2 := filepath.Join(tmpDir, "file321.txt")

		err := os.WriteFile(file1, []byte("content1"), 0o600)
		if err != nil {
			require.NoError(t, err)
		}

		err = os.WriteFile(file2, []byte("content2"), 0o600)
		if err != nil {
			require.NoError(t, err)
		}

		result, err := read.FindFilesByPattern(filepath.Join(tmpDir, "*.txt"))
		require.NoError(t, err)
		require.Len(t, result, 2)
	})

	t.Run("No Files Found", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := read.FindFilesByPattern(filepath.Join(tmpDir, "*.gogo"))
		assert.ErrorIs(t, err, domain.ErrFinding)
	})
}

func TestReadLines(t *testing.T) {
	readFile := &reader.FileReader{}
	readURL := &reader.URLReader{}

	t.Run("Read Lines from File", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "file.txt")
		err := os.WriteFile(filePath, []byte("line1\nline2\nline3\n"), 0o600)
		require.NoError(t, err)

		config := &domain.InputConfig{Path: filePath}

		lines, err := readFile.ReadLines(config)
		require.NoError(t, err)

		var collectedLines []string

		var collectedNames []string

		filesUsed := make(map[string]struct{})

		var wg sync.WaitGroup

		mutex := &sync.Mutex{}

		for i := 0; i < 5; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				for line := range lines {
					mutex.Lock()

					data := strings.Split(line, "$")
					if _, ok := filesUsed[data[0]]; !ok {
						collectedNames = append(collectedNames, data[0])
						filesUsed[data[0]] = struct{}{}
					}

					collectedLines = append(collectedLines, data[1])
					mutex.Unlock()
				}
			}()
		}

		wg.Wait()
		assert.ElementsMatch(t, []string{"file.txt"}, collectedNames)
		assert.ElementsMatch(t, []string{"line1\n", "line2\n", "line3\n"}, collectedLines)
	})

	t.Run("Read Lines from URL", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/test.txt", r.URL.Path)
			w.WriteHeader(http.StatusOK)

			_, err := w.Write([]byte("mock file content\n"))
			if err != nil {
				return
			}
		}))
		defer server.Close()
		config := &domain.InputConfig{Path: server.URL + "/test.txt"}

		lines, err := readURL.ReadLines(config)
		require.NoError(t, err)

		var collectedLines []string

		var collectedNames []string

		filesUsed := make(map[string]struct{})

		var wg sync.WaitGroup

		mutex := &sync.Mutex{}

		for i := 0; i < 5; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				for line := range lines {
					mutex.Lock()

					data := strings.Split(line, "$")
					if _, ok := filesUsed[data[0]]; !ok {
						collectedNames = append(collectedNames, data[0])
						filesUsed[data[0]] = struct{}{}
					}

					collectedLines = append(collectedLines, data[1])
					mutex.Unlock()
				}
			}()
		}

		wg.Wait()
		assert.ElementsMatch(t, []string{"test.txt(from url)"}, collectedNames)
		assert.ElementsMatch(t, []string{"mock file content\n"}, collectedLines)
	})

	t.Run("Invalid URL", func(t *testing.T) {
		config := &domain.InputConfig{Path: "/nonexistent/file.txt"}
		_, err := readURL.ReadLines(config)
		assert.Error(t, err)
	})
}

func TestProcessURL(t *testing.T) {
	read := &reader.URLReader{}

	t.Run("Valid URL", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/test.txt", r.URL.Path)
			w.WriteHeader(http.StatusOK)

			_, err := w.Write([]byte("mock file content"))
			if err != nil {
				return
			}
		}))
		defer server.Close()

		result, name, err := read.ProcessURL(server.URL + "/test.txt")

		require.NoError(t, err)
		require.NotNil(t, result)
		data, err := io.ReadAll(result)
		require.Equal(t, []byte("mock file content"), data)
		require.NoError(t, err)
		require.Equal(t, name, "test.txt")
	})

	t.Run("Invalid URL", func(t *testing.T) {
		_, _, err := read.ProcessURL("http://nonexistent.url/file.txt")
		assert.Error(t, err)
	})

	t.Run("Non-200 Status Code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		_, _, err := read.ProcessURL(server.URL + "/test.txt")
		assert.ErrorIs(t, err, domain.ErrDownload)
	})
}
