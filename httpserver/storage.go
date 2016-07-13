package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"
)

var addr = flag.String("addr", ":80", "Listen ip:port")
var dir = flag.String("dir", ".", "Run on dir")

var fs *LocalFs

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	fs = &LocalFs{RootPath: *dir}

	http_server(*addr)
}

func http_server(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", methodRouter)

	s := http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    time.Second * 60,
		WriteTimeout:   time.Second * 60,
		MaxHeaderBytes: 1024 * 4,
	}

	log.Fatalln(s.ListenAndServe())
}

func methodRouter(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := recover()
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			log.Printf("%s %s 500 %v - %s\n", r.Method, r.URL.Path, err, r.RemoteAddr)
		}
	}()

	switch r.Method {
	case "GET":
		fallthrough
	case "HEAD":
		sendFile(w, r)
	case "PUT":
		saveFile(w, r)
	case "DELETE":
		deleteFile(w, r)
	default:
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		log.Println(r.Method + " " + r.URL.Path + " 405 Method Not Allowed " + r.RemoteAddr)
	}
}

//读取文件
func sendFile(w http.ResponseWriter, r *http.Request) {
	file := path.Clean(r.URL.Path)

	f, err := fs.Open(file)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		log.Println(r.Method + " " + file + " " + err.Error())
		http.NotFound(w, r)
		return
	}

	if d.IsDir() {
		http.NotFound(w, r)
		return
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
}

//保存文件
func saveFile(w http.ResponseWriter, r *http.Request) {
	file := path.Clean(r.URL.Path)
	dir := path.Dir(file)
	if dir != "/" {
		_, err := fs.Stat(dir)
		if os.IsNotExist(err) {
			fs.MkdirAll(dir, os.ModePerm)
		}
	}

	f, err := fs.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0777)
	if err != nil {
		http.Error(w, "Fail "+err.Error(), 500)
		log.Println("Put Fail " + file + " " + r.RemoteAddr + " " + err.Error())
		return
	}
	defer f.Close()

	_, err = io.Copy(f, r.Body)
	if err != nil {
		http.Error(w, "Fail "+err.Error(), 500)
		log.Println("Put Fail " + file + " " + r.RemoteAddr + " " + err.Error())
		return
	}

	w.Write([]byte("Success"))
}

//删除文件
func deleteFile(w http.ResponseWriter, r *http.Request) {
	file := path.Clean(r.URL.Path)

	err := fs.Remove(file)
	if err != nil {
		http.Error(w, "Fail "+err.Error(), 500)
		log.Println("Delete Fail " + file + " " + r.RemoteAddr + " " + err.Error())
		return
	}

	w.Write([]byte("Success"))
}
