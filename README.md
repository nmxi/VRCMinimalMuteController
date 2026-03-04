# VRCMinimalMuteController (ミュトコン)

<img width="402" height="69" alt="VRCMinimalMuteController_TaskBarImage" src="https://github.com/user-attachments/assets/5e4ad8bc-cd61-4739-a96d-2c23cff7daea" />

Windows の通知領域に常駐し、ダブルクリックまたは登録したショートカットで OSC を送信する最小構成の Go アプリです。

<img width="432" height="241" alt="VRCMinimalMuteController_SetupShortcut" src="https://github.com/user-attachments/assets/4e700075-81d0-4c82-9334-34136816e126" />

グローバルショートカットの設定が可能なので、VRChatのウィンドウがアクティブでなくてもマイクミュートの切り替え操作が可能です。

## 主な機能

- トレイ常駐
- `/input/Voice` に `0 -> 1 -> 0` を送信
- 送信間隔は 32ms
- スタートアップ有効化 / 無効化
- グローバルショートカットの登録 / 削除
- ショートカット設定画面で、修飾キーは `Ctrl` / `Shift` / `Alt` のチェックボックス、入力キーはキー入力またはプルダウンから選択可能
- 多重起動防止

## ショートカットの制約

- `Pause`、`Break`、`Home`、`End`、`PageUp`、`PageDown`、`Insert`、`Delete`、`F1-F24` は単独キーで登録できます。
- それ以外のキーは単独登録できません。`Ctrl` / `Shift` / `Alt` のいずれか1つ以上を組み合わせてください。
- `LWin` と `RWin` は登録できません。
