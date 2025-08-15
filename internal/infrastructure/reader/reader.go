package reader

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/4domm/ngxstat/internal/domain"
)

type FileReader struct {
}

func (fr *FileReader) ReadLines(inputConfig *domain.InputConfig) (lines chan string, err error) {
	var data []string
	data, err = fr.FindFilesByPattern(inputConfig.Path)

	if err != nil {
		return nil, err
	}

	lines = make(chan string)

	go func() {
		defer close(lines)

		for _, path := range data {
			file, err := os.Open(path)
			if err != nil {
				continue
			}

			reader := bufio.NewReader(file)

			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}

					continue
				}
				lines <- filepath.Base(file.Name()) + "$" + line
			}

			file.Close()
		}
	}()

	return lines, nil
}
func (fr *FileReader) FindFilesByPattern(pattern string) ([]string, error) {
	startPath := fr.getStartPath(pattern)

	if !strings.ContainsAny(pattern, "*?") {
		return fr.findSingleFile(pattern)
	}

	return fr.findFilesByPattern(startPath, pattern)
}

func (fr *FileReader) findSingleFile(pattern string) ([]string, error) {
	var data []string

	if _, err := os.Stat(pattern); err == nil {
		data = append(data, pattern)
	} else {
		return nil, domain.ErrFinding
	}

	return data, nil
}
func (fr *FileReader) findFilesByPattern(startPath, pattern string) ([]string, error) {
	var data []string

	err := filepath.WalkDir(startPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		matched, _ := filepath.Match(pattern, path)

		if matched && !d.IsDir() {
			data = append(data, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, domain.ErrFinding
	}

	return data, nil
}

func (fr *FileReader) getStartPath(pattern string) string {
	wildcardIndex := strings.IndexAny(pattern, "*?")
	if wildcardIndex == -1 {
		return filepath.Dir(pattern)
	}

	return filepath.Dir(pattern[:wildcardIndex])
}
