package humaystorage

import (
	"encoding/json"
	"errors"
	"os"
)

func (s *MemStorage) Save() error {
	buf, err := json.Marshal(s)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(s.storageFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	n, err := file.Write(buf)
	if err != nil {
		return err
	}

	if n == 0 {
		return errors.New("nothing is saved")
	}

	return nil
}

func (s *MemStorage) Restore(filePath string) error {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, s)
}
