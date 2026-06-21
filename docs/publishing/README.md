# Publishing Plan

公開は保留中。

ユーザー確認済みの方針:

- 機能要件を全部満たすまで release しない。
- QA が通るまで GitHub Pages / GitHub Release を更新しない。
- `kdevelopk.pages.dev` は作品一覧サイトなので、MofuMouse 単体ページとして直接デプロイしない。

有効化手順:

1. 全機能要件とアセット QA を完了する。
2. `go test ./...` と `scripts/build.ps1` を通す。
3. 画面表示 QA と x86/x64 起動確認を通す。
4. `docs/publishing/*.workflow.yml` を `.github/workflows/` へコピーする。
5. GitHub repository / Pages / Release を設定する。
6. 作品一覧へ載せる場合は `D:\開発\works-gallery` 側で作品カードを更新し、`kdevelopk.pages.dev` にデプロイする。
