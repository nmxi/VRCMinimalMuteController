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
- グローバルショートカット登録 / 削除
