package main

//TODO chの型とファイル作成イベントをコンテナイベントに読み替える方法を学ぶ
type Watcher struct {
	eventCh chan string
	path    string
}

func NewWatcher(eventCh chan string, path string) *Watcher {
	return &Watcher{
		eventCh: eventCh,
		path:    path,
	}
}

func (w *Watcher) Start() {

}

func ProcessEvent() {

}
