package web

import (
	"encoding/json"
	"hahajing/com"
	"hahajing/kad"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"golang.org/x/net/websocket"
)

const (
	kadSearchWaitingTime = 10
)

// webError is for user browser.
type webError struct {
	Error string
}

// Web x
type Web struct {
	searchReqCh chan *kad.SearchReq

	homeTemplate *template.Template
}

// Start x
func (we *Web) Start(searchReqCh chan *kad.SearchReq) {
	we.searchReqCh = searchReqCh

	// HTML page
	path := com.GetConfigPath()
	tmpl, err := template.ParseFiles(path + "/config/web/home.html")
	if err != nil {
		log.Panic("Home page failed!")
	}
	we.homeTemplate = tmpl

	// at last start sever
	we.startServer()
}

func (we *Web) readSearchInput(ws *websocket.Conn) ([]string, string) {
	// read
	msg := make([]byte, 1024)
	n, err := ws.Read(msg)
	if err != nil {
		return nil, "读取数据错误，请重试！"
	}

	// add new search
	ip, _, _ := net.SplitHostPort(ws.Request().RemoteAddr)

	// parse
	text := strings.TrimSpace(string(msg[:n]))
	com.HhjLog.Infof("New user(%s) input: %s", ip, text)

	keywords := strings.Split(text, " ")
	if keywords == nil {
		return nil, "没有搜索关键字，请重新输入！"
	}

	return keywords, ""
}

func (we *Web) writeError(ws *websocket.Conn, errStr string) {
	data, _ := json.Marshal(&webError{Error: errStr})
	ws.Write(data)
}

func (we *Web) send2Kad(ws *websocket.Conn, keywords []string) {
	respCh := make(chan *kad.SearchResp, kad.SearchRespChSize)
	searchReq := kad.SearchReq{RespCh: respCh, Keywords: keywords}
	we.searchReqCh <- &searchReq

	// waiting result from KAD
	found := false
	for {
		select {
		case pSearchResp := <-respCh:
			for _, fileLink := range pSearchResp.FileLinks {
				found = true
				ws.Write(fileLink.ToJSON())
			}
		case <-time.After(kadSearchWaitingTime * time.Second):
			if !found {
				we.writeError(ws, "搜索超时，请重试！")
			}
			return
		}
	}
}

func (we *Web) searchHandler(ws *websocket.Conn) {
	// read user input from network
	keywords, errStr := we.readSearchInput(ws)
	if keywords == nil {
		we.writeError(ws, errStr)
		return
	}

	// send to KAD
	if keywords == nil || len(keywords) <= 0 {
		we.writeError(ws, "关键词为空，请重新输入！")
		return
	}

	we.send2Kad(ws, keywords)
}

func (we *Web) homeHandler(w http.ResponseWriter, r *http.Request) {
	homeData := &HomeData{Host: "ws://" + r.Host + "/search"}
	err := we.homeTemplate.Execute(w, homeData)
	if err != nil {
		com.HhjLog.Criticalf("Execute template failed: %s", err)
	}
}

func (we *Web) startServer() {
	com.HhjLog.Info("Web Server is running...")

	http.HandleFunc("/", we.homeHandler)
	http.Handle("/search", websocket.Handler(we.searchHandler))

	var err error
	if len(os.Args) > 1 && os.Args[1] == "server" {
		err = http.ListenAndServe(":80", nil)
	} else {
		err = http.ListenAndServe(":66", nil)
	}
	if err != nil {
		com.HhjLog.Panic("Start Web Server failed: ", err)
	}
}
