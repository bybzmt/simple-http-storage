package main

import (
	"os"
	"path"
	"time"
)

type LocalFs struct {
	RootPath string
}

func (l *LocalFs) Init(rootPath string) *LocalFs {
	l.RootPath = rootPath
	return l
}

func (l *LocalFs) Chmod(name string, mode os.FileMode) error {
	_name := path.Join(l.RootPath, name)
	return os.Chmod(_name, mode)
}

func (l *LocalFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	_name := path.Join(l.RootPath, name)
	return os.Chtimes(_name, atime, mtime)
}

func (l *LocalFs) Mkdir(name string, perm os.FileMode) error {
	_name := path.Join(l.RootPath, name)
	return os.Mkdir(_name, perm)
}

func (l *LocalFs) MkdirAll(pathName string, perm os.FileMode) error {
	_path := path.Join(l.RootPath, pathName)
	return os.MkdirAll(_path, perm)
}

func (l *LocalFs) Remove(name string) error {
	_name := path.Join(l.RootPath, name)
	return os.Remove(_name)
}

func (l *LocalFs) RemoveAll(pathName string) error {
	_path := path.Join(l.RootPath, pathName)
	return os.RemoveAll(_path)
}

func (l *LocalFs) Rename(oldpath, newpath string) error {
	_oldpath := path.Join(l.RootPath, oldpath)
	_newpath := path.Join(l.RootPath, newpath)
	return os.Rename(_oldpath, _newpath)
}

func (l *LocalFs) Truncate(name string, size int64) error {
	_name := path.Join(l.RootPath, name)
	return os.Truncate(_name, size)
}

func (l *LocalFs) Create(name string) (file *os.File, err error) {
	_name := path.Join(l.RootPath, name)
	return os.Create(_name)
}

func (l *LocalFs) Open(name string) (file *os.File, err error) {
	_name := path.Join(l.RootPath, name)
	return os.Open(_name)
}

func (l *LocalFs) OpenFile(name string, flag int, perm os.FileMode) (file *os.File, err error) {
	_name := path.Join(l.RootPath, name)
	return os.OpenFile(_name, flag, perm)
}

func (l *LocalFs) Stat(name string) (fi os.FileInfo, err error) {
	_name := path.Join(l.RootPath, name)
	return os.Stat(_name)
}

func (l *LocalFs) Lstat(name string) (fi os.FileInfo, err error) {
	_name := path.Join(l.RootPath, name)
	return os.Lstat(_name)
}
