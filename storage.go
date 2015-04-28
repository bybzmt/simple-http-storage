package main

import (
	"net/http"
	"flag"
	"os"
	"io"
	"log"
	"fmt"
	"time"
	"runtime"
	flog "github.com/bybzmt/golang-filelog"
	netfs "github.com/bybzmt/golang-netfs"
)

var addr = flag.String("addr", ":7001", "Listen ip:port")
var http_addr = flag.String("http", "", "HTTP Listen. Defult disabled")
var dir = flag.String("dir", ".", "Run on dir")

var log_file = flag.String("log_file", "<stderr>", "log filename")
var log_priority = flag.String("log_priority", "local0:info", "Log Priority")
var log_prefix = flag.String("log_prefix", "storage", "log Prefix")

var fs netfs.FileSystem

var slog flog.Writer

func main() {
	flag.Parse()

	init_log()

	runtime.GOMAXPROCS(runtime.NumCPU())

	fs = &netfs.LocalFs{RootPath:*dir}

	if *http_addr != "" {
		go http_server(*http_addr)
	}

	netfs.Listen(*addr, fs)
}

func init_log() {
	var err error
	slog, err = flog.New(*log_file, *log_priority, *log_prefix)
	if err != nil {
		log.Fatalln(err)
	}
}

func http_server(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", methodRouter)

	s := http.Server{
		Addr: addr,
		Handler:mux,
		ReadTimeout: time.Second * 60,
		WriteTimeout: time.Second * 60,
		MaxHeaderBytes: 1024 * 4,
	}

	err := s.ListenAndServe()
	slog.Emerg(err.Error())
}

func methodRouter(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := recover()
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			slog.Err(fmt.Sprintf("%s %s 500 %v - %s", r.Method, r.URL.Path, err, r.RemoteAddr))
		}
	}()

	switch (r.Method) {
	case "GET" : fallthrough
	case "HEAD" :
		sendFile(w, r)
	case "PUT" :
		saveFile(w, r)
	case "DELETE" :
		deleteFile(w, r)
	default:
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		slog.Info(r.Method + " " + r.URL.Path + " 405 Method Not Allowed " + r.RemoteAddr)
	}
}

//读取文件
func sendFile(w http.ResponseWriter, r *http.Request) {
	f, err := fs.Open(r.URL.Path)
	if err != nil {
		slog.Info(r.Method + " " + r.URL.Path + " 404 Not Found")
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		slog.Info(r.Method + " " + r.URL.Path + " " + err.Error())
		http.NotFound(w, r)
		return
	}

	if d.IsDir() {
		slog.Info(r.Method + " " + r.URL.Path + " 404 Not Found")
		http.NotFound(w, r)
		return
	}

	slog.Info(r.Method + " " + r.URL.Path + " 200 Ok")

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
}

//保存文件
func saveFile(w http.ResponseWriter, r *http.Request) {
	f, err := fs.OpenFile(r.URL.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		http.Error(w, "Fail " + err.Error(), 500)
		slog.Notice("Put Fail" + r.URL.Path + " " + r.RemoteAddr + " " + err.Error())
		return
	}
	defer f.Close()

	_, err = io.Copy(f, r.Body)
	if err != nil {
		http.Error(w, "Fail " + err.Error(), 500)
		slog.Notice("Put Fail" + r.URL.Path + " " + r.RemoteAddr + " " + err.Error())
		return
	}

	slog.Info("Put " + r.URL.Path + " 200 Ok")

	w.Write([]byte("Success"))
}

//删除文件
func deleteFile(w http.ResponseWriter, r *http.Request) {
	err := fs.Remove(r.URL.Path)
	if err != nil {
		http.Error(w, "Fail " + err.Error(), 500)
		slog.Info("Delete Fail " + r.URL.Path + " " + r.RemoteAddr + " " + err.Error())
		return
	}

	slog.Info("Delete " + r.URL.Path + " 200 Ok")

	w.Write([]byte("Success"))
}
