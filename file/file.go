package file

import (
	"io/ioutil"
	"os"
)

type FileEvent struct {
	Filepath    string
	FileCreated bool
}

func SearchDir(dir string, filter func(filepath string) bool) ([]string, error) {
	var (
		fileInfos []os.FileInfo
		err       error
	)
	result := make([]string, 0, 256)
	if fileInfos, err = ioutil.ReadDir(dir); err != nil {
		return nil, err
	} else {
		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() {
				var filepaths []string
				if filepaths, err = SearchDir(fileInfo.Name(), filter); err != nil {
					return nil, err
				}
				result = append(result, filepaths...)
			} else {
				if !filter(fileInfo.Name()) {
					result = append(result, fileInfo.Name())
				}
			}
		}
	}
	return result, nil
}
