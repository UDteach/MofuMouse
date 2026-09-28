# MofuMouse

MofuMouse は、チューチューマウスの現代版を目指す Windows 用デグーアシスタントです。

現在の MVP は Go + Win32 API で、カーソル真横に ImageGen 生成デグー overlay を表示し、Wails + React + Fluent UI の設定画面から表示、歩行/待機アクション、1px 移動、クイックキーを切り替えられます。

## 現在できること

- カーソル横にデグーを表示
- 32px 前後の小型表示。既定では名前ラベルを常時表示しません。
- カーソル移動中は補間追従し、移動方向に応じて左右へ回り込みます。
- カーソル移動中は `walk` / `walk2` / `run` の複数フレームで歩行感を出します。
- 1px 移動が有効な時は guard ポーズで状態を示します。
- 一定時間カーソルが止まると、かじる/毛づくろい/ぺちぺち/匂い嗅ぎ/眠る/掘る/転がる idle action を順番に出します。
- トレイのクイックキーから、検索 / コピー / 全選択 / 新しいタブの固定ショートカットだけを実行できます。
- トレイのクイックスクロールから、今のウィンドウを少しだけ上下に動かせます。
- Windows 起動時に MofuMouse を自動で始める設定を選べます。既定はオフです。
- GitHub Releases を確認し、このPC向けの x86/x64 更新ファイルをローカル更新キャッシュへダウンロードできます。トレイから確認付きで適用して再起動でき、CLI で本体置換と rollback も実行できます。
- click-through / topmost / no-activate overlay
- モダン設定画面
  - Wails + React + Fluent UI
  - Home / Companion / Jiggle / Appearance / Updates / Advanced のタブ
  - デグーの名前
  - デグーの大きさ
  - 毛色
  - カーソル横表示
  - クイックキーの固定allowlist
  - クイックスクロール
  - スクリーンセーバー防止の 1px 移動
  - 無操作時のしぐさ
  - 画面端で反対側へ移動
  - Windows 起動時に始める
- トレイメニュー
  - 設定を開く
  - デグーをカーソル横に表示 ON/OFF
  - スクリーンセーバー防止の 1px 移動 ON/OFF
  - 止まっている時のしぐさ ON/OFF
  - Windows起動時にMofuMouseを始める ON/OFF
  - アップデート確認。新しい版があり、このPC向けファイルが見つかった場合はダウンロードし、確認後に適用して再起動できます。
  - 毛色: ノーマル / グレー / ダークブラウン / クリーム / ホワイト / ぶち模様
  - クイックキー: 検索 / コピー / 全選択 / 新しいタブ
  - クイックスクロール: 上へ少し / 下へ少し
  - デスクトップ表示 / スタートメニュー / ブラウザ戻る・進む
  - 終了
- 名前設定と JSON 設定保存
- Windows x86 / x64 ビルド
  - 本体は x86 / x64
  - モダン設定画面は x64。x86 または未配置時は Win32 設定画面へフォールバックします。
- `--smoke` による短時間起動確認

## 画面の設定

- `デグーの名前`: トレイの状態表示に使う名前です。画面に常時出すかどうかは `デグーの名前を画面に出す` で選びます。
- `デグーの大きさ`: カーソル横に出るデグーの表示サイズです。見づらい時だけ大きくします。
- `毛色`: 見た目だけの設定です。あとから何度でも変えられます。
- `カーソル横にデグーを表示する`: マウスカーソルのすぐ横にデグーを出します。カーソルの移動方向に合わせて左右へ回り込みます。
- `クイックキー`: トレイから使う固定ショートカットです。`Ctrl+F` / `Ctrl+C` / `Ctrl+A` / `Ctrl+T` だけを許可し、削除・送信・閉じる系は入れません。
- `クイックスクロール`: トレイから今のウィンドウを少しだけ上下にスクロールします。連続自動スクロールはしません。
- `スクリーンセーバー防止の1px移動`: しばらく触っていない時だけ、マウスを 1px だけ戻して動かします。
- `止まっている時にしぐさを出す`: かじる、毛づくろい、ぺちぺちなどの待機アクションを出します。
- `画面端で反対側へ移動する`: カーソルが画面端に触れた時、反対側へ移動します。慣れてから使う設定です。
- `Windows起動時に始める`: サインイン時に `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` へ登録した MofuMouse を起動します。管理者権限は不要で、オフにすると登録を削除します。

## アセット方針

正式なデグー素材は必ず ImageGen で生成します。

overlay の runtime sprite は ImageGen で作った個別 PNG を採用します。GDI の簡易描画は、sprite 読み込み失敗時のフォールバックとしてだけ残します。

パースと画風を安定させるため、正式 runtime 素材は sprite sheet ではなく 1 枚ずつ生成します。シート切り抜きは style reference / candidate 扱いです。
現在の runtime 採用素材は、全 6 毛色の `idle` / `walk` / `walk2` / `run` / `sniff`、agouti の `return` / `guard` / `patpat` / `nibble` / `groom` / `sleepy` / `dig` / `roll` です。毛色別アクション素材が未生成の pose は、同じ毛色の idle へフォールバックします。追加ポーズと毛色は `.codex/tasks/mofumouse-mvp-ledger.md` のアセット制作台帳で管理します。

## リリース方針

まだリリースしません。機能要件とアセット制作が揃い、QA が通ってから GitHub Pages / Cloudflare Pages / GitHub Release を有効化します。公開用 workflow は `docs/publishing/` にテンプレートとして置いてあります。

## ビルド

```powershell
.\scripts\build.ps1
```

成果物:

- `dist\mofumouse-x86.exe`
- `dist\mofumouse-x64.exe`
- `dist\mofumouse-control-x64.exe`

## 手動確認

```powershell
go test ./...
go build -ldflags="-H=windowsgui" -o mofumouse.exe ./cmd/mofumouse
.\mofumouse.exe --smoke
```

通常起動:

```powershell
.\dist\mofumouse-x64.exe
```

終了はトレイメニューの「終了」から行います。

設定画面の起動確認:

```powershell
.\dist\mofumouse-x64.exe --open-settings
.\dist\mofumouse-x64.exe --open-settings --legacy-settings
```

歩行フレームQA:

```powershell
go test ./assets ./internal/winapp
```

display / DPI QA:

```powershell
.\scripts\verify_display_qa.ps1 -AppPath .\dist\mofumouse-x64.exe
.\scripts\verify_display_qa.ps1 -AppPath .\dist\mofumouse-x86.exe
```

自動起動QA:

```powershell
.\scripts\verify_autostart_qa.ps1
```

アップデートQA:

```powershell
.\scripts\verify_update_qa.ps1
```

クイックキーQA:

```powershell
.\scripts\verify_quickkey_qa.ps1
```

クイックスクロールQA:

```powershell
.\scripts\verify_quickscroll_qa.ps1
```

return pose QA:

```powershell
.\scripts\capture_return_pose_qa.ps1 -ExePath .\dist\mofumouse-x64.exe -ConfigPath .\.codex\qa\qa-target-move-return-settings.json -OutputPath .\.codex\qa\overlay-return-pose.png
```

GUI QA は前面ウィンドウとカーソルを使うため、並列ではなく順番に実行します。

## 設計メモ

- 要件: `docs/degu-mofumouse-requirements.md`
- ちゅーちゅーマウス機能ロードマップ: `docs/chuchu-mouse-feature-roadmap.md`
- 設計: `docs/degu-mofumouse-design.md`
- 技術選定 ADR: `docs/adr/0001-choose-go-win32-for-mvp.md`
- 設定UI ADR: `docs/adr/0002-modern-settings-control-center.md`
- 作業台帳: `.codex/tasks/mofumouse-mvp-ledger.md`
