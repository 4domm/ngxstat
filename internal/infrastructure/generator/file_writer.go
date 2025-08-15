package generator

import "os"

type FileWriter struct {
}

func (fw FileWriter) CreateFile(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}

	return file, nil
}
