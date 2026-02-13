package controller

import 	(
	"net/http"
	"strings"
)

type StaticFileHandler struct {
	fileServer http.Handler
}

func (sfh *StaticFileHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/webgl") {
		// Unityの圧縮ファイル（.gzや.br）に対する適切なヘッダー付与
		if strings.HasSuffix(r.URL.Path, ".gz") {
			w.Header().Set("Content-Encoding", "gzip")
		} else if strings.HasSuffix(r.URL.Path, ".br") {
			w.Header().Set("Content-Encoding", "br")
		}

		// .wasmファイルには正しいMIMEタイプが必要
		if strings.HasSuffix(r.URL.Path, ".wasm.gz") {
			w.Header().Set("Content-Type", "application/wasm")
		}
	}

	sfh.fileServer.ServeHTTP(w, r)
}

func NewStaticFileHandler(rootDir http.FileSystem) *StaticFileHandler {
	fileServer := http.FileServer(rootDir)
	return &StaticFileHandler{fileServer: fileServer}
}