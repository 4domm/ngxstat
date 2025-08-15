package reader

import (
	"bufio"
	"io"
	"net/http"
	"path/filepath"

	"github.com/4domm/ngxstat/internal/domain"
)

const ValidStatusCode = 200

type URLReader struct {
}

func (ur *URLReader) ReadLines(inputConfig *domain.InputConfig) (lines chan string, err error) {
	data, name, err := ur.ProcessURL(inputConfig.Path)

	if err != nil {
		return nil, err
	}

	lines = make(chan string)

	go func() {
		defer close(lines)

		reader := bufio.NewReader(data)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}

				continue
			}
			lines <- name + "(from url)$" + line
		}

		data.Close()
	}()

	return lines, nil
}

func (ur *URLReader) ProcessURL(path string) (io.ReadCloser, string, error) {
	req, err := http.NewRequest(http.MethodGet, path, http.NoBody)

	if err != nil {
		return nil, "", err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode != ValidStatusCode {
		resp.Body.Close()
		return nil, "", domain.ErrDownload
	}

	return resp.Body, filepath.Base(path), nil
}
