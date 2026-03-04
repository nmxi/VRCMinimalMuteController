//go:build windows

package main

import _ "embed"

// トレイアイコンは実行ファイルに埋め込んで、単体配布できるようにする。
//
//go:embed MuteIcon.ico
var embeddedTrayIcon []byte
