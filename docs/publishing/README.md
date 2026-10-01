# 更新版の公開

2026-10-01のユーザー依頼で、デグー全10色・10種26姿の `v0.1.0-preview.2` を公開する。ユーザーは旧Macの対象をmacOS 12までと指定した。今回の公開は他種40色の制作完了を意味しない。

AnimalsDesktopForRealの実構成を確認した。Electron43.7.5、Mac最小12.0、Apple Silicon／Intel DMG・ZIP、Windows x64 NSIS・ZIP、Macアドホック署名が参照元。MofuMouseの通常版は既存Electron44.4.5／Mac最小13.0を保持し、別名 `macos12` の互換版をElectron43.7.5／Mac最小12.0で作る。同一のレビュー済み画像を使う。Mac実機・Montereyでの起動は未確認と明記する。

公開入口は `.github/workflows/release.yml` の手動起動だけ。pushやtagでは自動起動しない。Actions節約のため、新しいrunはまず1回。失敗時は自動再実行せず原因と追加runの要否を報告する。

1. appの固定3,416PNG、27テスト、Windows配布物と実表示を確認する。
2. `node scripts/build-web-demo.mjs`、`node scripts/verify_preview_page.mjs`、`node scripts/stage-preview-site.mjs` を実行し、desktop/mobileとデモ操作を確認する。
3. package／lockfile／ページ／notesを同じ版へそろえ、レビュー可能なコミットをmainへ反映する。
4. `gh workflow run release.yml --repo UDteach/MofuMouse --ref main -f tag=v0.1.0-preview.2` を1回実行してrun IDを記録する。
5. 同じrunでWin2・Mac通常4・Mac互換4の10配布物を作る。Macは4appのCPU・最小OS・署名とホストCPUの実行環境版を確認する。`verify-release-build.mjs` は同梱ソース／画像／出力hashを検証し、通常2＋互換1の3つのbuild-infoを作る。
6. `scripts/publish-release.mjs` が同一コミット・カタログ・枚数・実行環境・最小OS・全10ファイルを照合し、GitHub prereleaseとPagesを同じrunで更新する。SHA256SUMSは10配布物＋3reportを含む。
7. Releaseの14添付物、Pagesのデモ・DLリンク、OS別選択、実ダウンロードを確認する。初回警告・更新・削除・不具合報告の手順を公開ページへ載せる。

公開ページはindex.html、download.html、assets、tryだけ。制作台帳・QA・ソース画像履歴は配信しない。GitHub上の旧releaseは残す。旧 `publish-preview.mjs` はpreview.1の履歴契約で、今回のワークフローから呼ばない。

一次情報: [ElectronのOS変更](https://github.com/electron/electron/blob/main/docs/breaking-changes.md)、[サポート方針](https://www.electronjs.org/docs/latest/tutorial/electron-timelines)、[Apple初回起動](https://support.apple.com/ja-jp/102445)。互換版の将来更新はElectron43の公式サポート状況を再確認する。古いOS対応を理由にサポート終了した実行環境へ自動で下げない。
