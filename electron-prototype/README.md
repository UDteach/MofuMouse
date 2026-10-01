# MofuMouse Electron

Windows／Mac向けのカーソル追従アプリ。preview.2は10種・26姿、デグー全10色、レビュー済みPNG3,416枚を収録します。他種の追加40色は別の制作作業として継続中です。

## 操作

通知領域／メニューバーの動物アイコンで、1〜10匹・32／48／64／96px・一時停止／再開／終了を選びます。「1匹ずつ選ぶ」は各順番の種類と毛色を保存し、隠れた順番も保持します。「動物・毛色（全員）」で全員を変更します。既定10匹・48px、保存設定があればそれを使います。

終了キーはWindowsでCtrl+Alt+Shift+Q、MacでCommand+Option+Shift+Q。クリックを通す透明Canvasを画面ごとに1つ使い、カーソルのある画面へ隊列を移します。MacのSpaces・フルスクリーンや実モニタの抜き差しは追加実機確認が必要です。

## ビルド

Node.js24で `npm ci --no-audit --no-fund`、`npm run check`。起動は `npm start`。Windows上で `npm run dist:win` はx64 NSIS/ZIPを `release-build/` に作ります。ZIPは全体を展開しMofuMouse.exeを開きます。Node/Pythonは配布アプリの実行に不要です。

MacのビルドはMac上で行います。

```sh
npm run dist:mac
node scripts/verify-release-build.mjs
npm run dist:mac:monterey
node scripts/verify-release-build.mjs --monterey
```

通常版はElectron44.4.5／macOS最小13.0、互換版はElectron43.7.5／macOS最小12.0です。Apple Silicon／Intel別DMG/ZIP。通常版は `release-build/`、互換版は `release-build-macos12/`。アドホック署名の整合性を確認しますが、Developer ID署名・公証は未実施です。OSセキュリティの一括解除コマンドは配布しません。macOS11以前へは自動で古い実行環境を下げません。

Windowsコード署名は未実施。Mac実機の表示・操作、Monterey実機での起動は未確認です。OS別パッケージと初回起動は [DLページ](https://udteach.github.io/MofuMouse/download.html) に記載しています。

## 固定素材とWebデモ

`app/media/manifest.json` の3,416PNGを起動・梱包前後にSHA-256で照合します。ビルドは作業中の素材台帳を再取り込みしません。デグーの歩行は全10色30実コマ／556ms。色違いの待機は8コマ／4秒、アグーチ基準色は元の96コマ待機を保持します。他種の歩行コマ数は種類によって異なります。

表示倍率・接地位置はmotion単位の共通変換で、元PNGを変更しません。制作候補、source採用、runtime採用を区別します。新規採用は `integrate-reviewed-coat.py` のhash-bound receiptとレビューでstage→applyし、既存原本・旧出力・生成回数を保持します。

ルートの `node scripts/build-web-demo.mjs` は同じH96 PNG1,708枚と `app/motion.mjs` を `docs/try/` へコピーします。Web用カタログに内部制作記録を含めません。デモは選択中の種類のみデコードし、タブ非表示では描画を休み、動きを減らす設定では停止して開始します。

## 検証

30件の単体・配布契約テスト。Windows配布アプリでは全26姿を切り替え、全画像の読み込み・個別指定・混在3／10種・隠れた順番・連続選択・停止／再開・フォーカスを確認します。黒の個別smokeは30歩行／8待機の全実描画、4サイズ、overlay再作成を確認しました。Native検査はウィンドウの透過・noActivate属性と背後へのhitを読み取ります。診断カーソル軌跡を与え、OSカーソルは動かしません。smokeは独立プロファイルを使い、通常の設定を変更しません。

`verify-release-build.mjs` は全アーキテクチャのapp.asar内ソース・PNGと成果物hashを照合します。MacCIでは4appのCPU・最小OS・署名、ホストCPUの実行環境版も確認します。CIでの検証は実機の見た目の確認と別です。

Webはdesktop／390／320幅で表示・追従・混在・毛色／数／サイズ／背景・一時停止・FAQ・keyboard・通信失敗からの再試行を確認します。物理スマートフォン／SafariとMac実機はこのWindowsホストでは未検証です。

公開手順は [publishing](../docs/publishing/README.md)。`pack:*` とMac source kitは旧ローカル開発経路で、配布版はBuilderを使います。ユーザー設定は既存の `MofuMouseElectronPrototype` フォルダーを保持するためpreview.1から引き継げます。
