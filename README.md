# toyAdvisor

https://github.com/google/cadvisor

cAdvisorを参考にしたコンテナ監視ツール(おもちゃ)

・監視コンテナはdockerに絞る

・計測対象はcpu,memoryのみ

としています。

## 目標
・現在稼働しているコンテナを監視する

・ツール起動中に作られた新規コンテナも監視する

・ツールで得た値をdocker statsコマンドで出力される値と同じようにしてファイルに出力




