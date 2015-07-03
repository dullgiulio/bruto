package gen

import (
	"bufio"
	"os"
)

type Filelines []string

func MakeFilelines() Filelines {
	return Filelines(make([]string, 0))
}

func (f *Filelines) add(s string) {
	*f = append(*f, s)
}

func (f *Filelines) Load(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f.add(scanner.Text())
	}
	return scanner.Err()
}
