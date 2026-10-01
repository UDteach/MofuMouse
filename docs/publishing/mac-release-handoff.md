# preview.2 のMacローカル検証

現行の公開は `.github/workflows/release.yml` で行う。最新指示によりMac版も今回公開する。この文書はMacでローカル検証したい場合、または後日手動でMac配布物を追加する場合の手順。Mac実機の起動・表示・Montereyでの動作はまだ確認していない。

Node.js24とGitを使い、公開タグの新しい作業フォルダーで実行する。既存の変更を上書きしない。

```sh
git clone https://github.com/UDteach/MofuMouse.git MofuMouse-preview2
cd MofuMouse-preview2
git checkout v0.1.0-preview.2
cd electron-prototype
npm ci --no-audit --no-fund
npm run check
npm run dist:mac
npm run dist:mac:monterey
bash scripts/verify-mac-packages.sh
```

通常版はElectron44.4.5／最小macOS13.0で `release-build/`、Monterey版はElectron43.7.5／最小macOS12.0で `release-build-macos12/`。どちらもApple Silicon／Intel各DMG・ZIP、計8配布物。Montereyの最小OSは型を保持したYAML文字列 `'12.0'` を使う。検査はCPU・実際のInfo.plist最小OS・アドホック署名・ホストCPUのElectron実行・同梱ソース・3,416PNG・配布物hashを照合し、2つのbuild-infoを作る。失敗したまま配布しない。

実機では通常サイズでwalk／idle、種類変更、停止／再開、クリック透過、終了、保存設定を確認する。Monterey実機確認とCI構成検査は別。Developer ID署名・公証は未実施。

GitHub CLIを使う場合、検証済みMacを既存Releaseへ追加できる。まず公開済みWindowsの元情報を読み戻し、全10配布物を照合する。

```sh
cd ..
mkdir release-assets
gh release download v0.1.0-preview.2 --repo UDteach/MofuMouse --pattern '*win-x64*' --pattern 'build-info-win32.json' --pattern 'windows-reuse.json' --dir release-assets
cp electron-prototype/release-build/*.dmg electron-prototype/release-build/*mac-*.zip electron-prototype/release-build/build-info-darwin.json release-assets/
cp electron-prototype/release-build-macos12/*.dmg electron-prototype/release-build-macos12/*.zip electron-prototype/release-build-macos12/build-info-macos12.json release-assets/
export GITHUB_SHA="$(git rev-parse HEAD)"
export EXPECTED_TAG=v0.1.0-preview.2
node scripts/publish-release.mjs --check-only
```

15添付物の検査成功後に、8つのMac配布物・2つのMac build-info・更新したSHA256SUMSを追加する。Windowsの配布物や元コミットの記録は置き換えない。次の `--clobber` は既存チェックサムの更新にだけ使う。

```sh
gh release upload v0.1.0-preview.2 release-assets/*mac-*.dmg release-assets/*mac-*.zip release-assets/*macos12-*.dmg release-assets/*macos12-*.zip release-assets/build-info-darwin.json release-assets/build-info-macos12.json --repo UDteach/MofuMouse
gh release upload v0.1.0-preview.2 release-assets/SHA256SUMS.txt --repo UDteach/MofuMouse --clobber
```

Macが公開された後、`docs/release-status.json` のavailableProfilesへdarwin／macos12を移し、DLページの準備中表示を実際の8リンクに替え、確認状況とRelease本文を実際の結果に合わせて更新する。Web公開は `website.yml` の手動起動で、アプリを再ビルドしない。Actions回数の指定がある場合はその範囲内で行う。
