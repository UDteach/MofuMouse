# 先行版の公開

2026-09-28のユーザー依頼により、確認済み19姿のElectron先行版を公開する。全色・全30コマの完成条件は継続中で、今回の公開を全制作完了と混同しない。

以前のGo版の機能完了待ち方針と個別Pages/Releaseテンプレートは履歴。現行の実行入口は `.github/workflows/release.yml` の手動 `workflow_dispatch` だけ。mainへのpushやtagでは自動起動しない。旧 `pages.yml` は統合した。

1. 固定素材のhash、単体テスト、Windows配布物/実表示、ページのdesktop/mobile/リンクをローカル確認する。
2. package.json / lockfile / ページ / release notesの版を合わせる。今回の版は `v0.1.0-preview.1`。
3. mainへ変更を反映する。
4. `gh workflow run release.yml --repo UDteach/MofuMouse --ref main -f tag=v0.1.0-preview.1` を**1回だけ**実行し、run IDを記録する。
5. 同じrunでWindows NSIS/ZIP、Mac arm64/x64 DMG/ZIPを作り、素材と配布物hashを検証し、GitHub prereleaseと既存Pagesを更新する。公開ページへはindex.htmlとassetsだけを送る。
6. Releaseの6配布物、SHA256SUMS、2つのbuild-infoとページのリンクを確認する。MacのCIビルド/署名整合性チェックは実機での表示・操作確認と別。

ユーザーのActions節約指定により、失敗時も自動rerunや2回目dispatchをしない。ローカルで原因を特定し、追加runが必要な場合は実施せず状況を報告する。Artifacts保持は1日。公開対象は `https://udteach.github.io/MofuMouse/` とUDteach/MofuMouseの先行版のみで、作品一覧サイトを直接上書きしない。

Mac minimumSystemVersionは13.0。採用Electron44の公式変更履歴がmacOS12非対応を明記している: https://www.electronjs.org/docs/latest/breaking-changes/ 。参照したAnimalsDesktopForRealの旧Electron43/min12設定をそのまま流用しない。
