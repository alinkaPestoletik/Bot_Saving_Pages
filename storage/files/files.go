package files

import (
	"BotSavingPages/lib/e"
	"BotSavingPages/storage"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Storage struct {
	basePath string
}

const defaultPerm = 0774

var ErrNoSavedPages = errors.New("no saved pages")

func New(basePath string) Storage {
	return Storage{basePath: basePath}
}

func (storage Storage) Save(page *storage.Page) (err error) {
	defer func() { err = e.WrapIfErr("can't save page", err) }()

	flpath := filepath.Join(storage.basePath, page.UserName)

	if err := os.MkdirAll(flpath, defaultPerm); err != nil {
		return err
	}
	fName, err := fileName(page)
	if err != nil {
		return err
	}
	flpath = path.Join(flpath, fName)
	file, err := os.Create(flpath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return err
	}
	return nil
}

const msgNoSavedPages = "You have no saved pages 🙊"

func (s Storage) PickRandom(username string) (page *storage.Page, err error) {
	fpath := filepath.Join(s.basePath, username)
	files, err := os.ReadDir(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.ErrNoSavedPages
		}
	}
	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(files))

	file := files[n]

	return s.decodePage(filepath.Join(fpath, file.Name()))
}

func (s Storage) Remove(page *storage.Page) error {
	fileName, err := fileName(page)
	if err != nil {
		return e.Wrap("can't remove file", err)
	}

	flpath := filepath.Join(s.basePath, page.UserName, fileName)

	if err := os.Remove(flpath); err != nil {
		return e.Wrap(fmt.Sprintf("can't remove file %s", flpath), err)
	}

	return nil
}

func (s Storage) IsExists(page *storage.Page) (bool, error) {
	fileName, err := fileName(page)
	if err != nil {
		return false, e.Wrap("can't check if file exists", err)
	}

	flpath := filepath.Join(s.basePath, page.UserName, fileName)

	switch _, err = os.Stat(flpath); {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:

		return false, e.Wrap(fmt.Sprintf("can't check if file %s exists", flpath), err)
	}
	return true, nil
}

func (s Storage) decodePage(filePage string) (*storage.Page, error) {
	file, err := os.Open(filePage)
	if err != nil {
		return nil, e.Wrap("can't decode page", err)
	}
	defer func() { _ = file.Close() }()

	var p storage.Page

	if err := gob.NewDecoder(file).Decode(&p); err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	return &p, nil
}

func fileName(page *storage.Page) (string, error) {
	return page.Hash()
}
