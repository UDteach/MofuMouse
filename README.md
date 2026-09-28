# MofuMouse

写真風の小動物がマウスカーソルについてくる、Windows / Mac向けの先行版です。

[ダウンロードと使い方](https://udteach.github.io/MofuMouse/) · [v0.1.0-preview.1](https://github.com/UDteach/MofuMouse/releases/tag/v0.1.0-preview.1)

10種・19種類の姿を収録しています。1〜10匹を表示し、通知領域／メニューバーの「1匹ずつ選ぶ」からそれぞれの動物と色を指定できます。大きさは32 / 48 / 64 / 96px。マウスの速さに合わせて歩き、止まると待機します。動物はクリックを遮りません。

Windows 10/11 x64向けにインストーラーとZIP、macOS 13以降のApple Silicon / Intel向けにDMGとZIPを用意します。Windowsで基本動作を確認しています。Macの実機表示・操作、長時間性能は未確認です。Windowsコード署名、Apple Developer ID署名・公証は行っていません。

動物の寸法、毛色、歩行は調整中です。デグーのブルー／サンドは待機時に静止画を表示します。現在の先行版は、すべての動物が30コマにそろった版ではありません。

## 開発・ビルド

Node.js 24を使います。

```sh
cd electron-prototype
npm ci
npm run check
npm start
```

```sh
npm run dist:win
npm run dist:mac
```

Windowsの配布物はWindowsで、Macの配布物はMacで作ります。出力先は `electron-prototype/release-build/` です。動物素材は検証済みの固定PNG列を同梱し、起動時にもhashを確認します。制作中の素材をビルド時に自動で取り込みません。

[詳しい開発手順](electron-prototype/README.md) · [先行版の内容](docs/publishing/preview-release-notes.md) · [公開手順](docs/publishing/README.md)

この先行版はElectronで実装しています。以前のGo版のソースは保持しており、説明は [旧版README](docs/legacy-go-readme.md) に残しています。旧版の1px移動やショートカット補助機能は現先行版に含みません。
