package general

import (
	"encoding/csv"
	"os"
)

// Create excel :
func CreateExcel(records *[][]string, path string, filename string) error {
	// check the directory exists
	CreateFolder(path)

	// generate the csv file
	f, err := os.Create(path + filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	// calls Flush internally
	err = w.WriteAll(*records)
	if err != nil {
		return err
	}

	return nil
}
