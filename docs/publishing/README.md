# 更新版の公開

2026-10-01のユーザー依頼で、デグー全10色・10種26姿の `v0.1.0-preview.2` を公開する。ユーザーは旧Macの対象をmacOS 12までと指定した。今回の公開は他種40色の制作完了を意味しない。

AnimalsDesktopForRealの実構成を確認した。Electron43.7.5、Mac最小12.0、Apple Silicon／Intel DMG・ZIP、Windows x64 NSIS・ZIP、Macアドホック署名が参照元。MofuMouseの通常版は既存Electron44.4.5／Mac最小13.0を保持し、別名 `macos12` の互換版をElectron43.7.5／Mac最小12.0で作る。同一のレビュー済み画像を使う。Mac実機・Montereyでの起動は未確認と明記する。

2026-10-01の最新指定「いや、Mac版も公開しておいてくれ」により、Mac通常版・Monterey互換版も今回公開する。Windowsの成功済み配布物は再利用し、修正済みMac構成だけ追加1runでビルド・検証・公開する。これは先のMac担当引継ぎと追加run確認待ちを置き換える直接指示。旧版preview.1は残す。[Macでのローカル検証手順](mac-release-handoff.md)。失敗したまま公開せず、再失敗時に自動の反復実行はしない。

検証済みWindowsの直接公開は `scripts/publish-release.mjs --windows-only`。元Windows jobの成功、Git object一致、build-info、2配布物のサイズとhashを確認して元情報を保持する。Windows2配布物＋build-info＋windows-reuse＋SHA256SUMSの5添付物を公開する。Macの8配布物が欠けたまま全構成公開の検査を通すことはできない。

既存Releaseを確認してWebだけ更新する入口は `.github/workflows/website.yml`。画像・motion・ページ内リンク・実公開済みの配布物とchecksumを検証してPagesへ反映する。Nodeのみで動き、アプリのビルドは行わない。公開可能profileを `docs/release-status.json` へ記録し、準備中のMac新版を配布済みと表示しない。pushやtagでは自動起動しない。追加Actionsの回数制限は維持し、手動dispatchは指示された範囲内で実行する。

1. appの固定3,416PNG、配布契約を含む単体テスト、Windows配布物と実表示を確認する。
2. `node scripts/build-web-demo.mjs`、`node scripts/verify_preview_page.mjs`、`node scripts/stage-preview-site.mjs` を実行し、desktop/mobileとデモ操作を確認する。
3. package／lockfile／ページ／notesを同じ版へそろえ、レビュー可能なコミットをmainへ反映する。
4. `gh workflow run release.yml --repo UDteach/MofuMouse --ref main -f tag=v0.1.0-preview.2` を1回実行してrun IDを記録する。
5. 同じrunでWin2・Mac通常4・Mac互換4の10配布物を作る。Macは4appのCPU・最小OS・署名とホストCPUの実行環境版を確認する。`verify-release-build.mjs` は同梱ソース／画像／出力hashを検証し、通常2＋互換1の3つのbuild-infoを作る。
6. `scripts/normalize-release-assets.mjs` はArtifacts内のbuildディレクトリを保持し、配布ファイルを公開用へ集める。`scripts/publish-release.mjs` がソースコミット・カタログ・枚数・実行環境・最小OS・全10ファイルを照合し、GitHub prereleaseとPagesを同じrunで更新する。SHA256SUMSは10配布物＋3reportを含む。
7. Releaseの14添付物、Pagesのデモ・DLリンク、OS別選択、実ダウンロードを確認する。初回警告・更新・削除・不具合報告の手順を公開ページへ載せる。

Macだけの修正後に検証済みWindowsを再利用する場合は、追加runの承認を得て `windows_run` に元run IDを指定する。成功したWindows job、元コミット、app全体・build全体・package／lockfile・通常builder設定・素材準備スクリプトのGit objectがすべて同じことを確認する。build-infoの元コミットは書き換えず、`windows-reuse.json` に元run・元コミット・公開コミット・一致objectを添付する。公開側でもGit objectを再照合し、全10配布物のサイズとSHA-256を検証する。この場合は15添付物、SHA256SUMSは14ファイル。アプリやWindowsビルド入力が変わった場合は再利用を拒否する。

MontereyのOS設定は `electron-builder.monterey.yml` の文字列 `'12.0'` を使う。CLIの数値変換が起きるdotted overrideを使わない。macOS CIは実際のCPU・LSMinimumSystemVersionをログへ出し、厳密に照合する。

公開ページはindex.html、download.html、assets、tryだけ。制作台帳・QA・ソース画像履歴は配信しない。GitHub上の旧releaseは残す。旧 `publish-preview.mjs` はpreview.1の履歴契約で、今回のワークフローから呼ばない。

一次情報: [ElectronのOS変更](https://github.com/electron/electron/blob/main/docs/breaking-changes.md)、[サポート方針](https://www.electronjs.org/docs/latest/tutorial/electron-timelines)、[Apple初回起動](https://support.apple.com/ja-jp/102445)。互換版の将来更新はElectron43の公式サポート状況を再確認する。古いOS対応を理由にサポート終了した実行環境へ自動で下げない。
