写真風の小動物が、マウスカーソルについてくるMofuMouseの更新版です。

[**Webでおためし**](https://udteach.github.io/MofuMouse/try/) · [**ダウンロードと使い方**](https://udteach.github.io/MofuMouse/download.html) · [公式サイト](https://udteach.github.io/MofuMouse/)

![MofuMouseの10種の表示例](https://udteach.github.io/MofuMouse/assets/mofumouse-preview-ten-animals.png)

## 今回の更新

- **デグー全10色**を収録。アグーチ、ブルー、サンド、ブラック、ホワイト、アグーチパイド、ブルーパイド、サンドパイド、クリームパイド、ブラックパイド。
- デグーの色違いは全30コマの歩行と8コマの待機。preview.1で静止していたブルー／サンドも待機中に動きます。基準色の連続動画による待機は保持しています。
- 全体で**10種・26種類**。1〜10匹を選び、1匹ずつ動物と毛色を設定。32／48／64／96pxに対応します。
- Webデモで毛色・数・サイズ・背景を変更でき、10種を並べてページ内の追従を試せます。
- **macOS 12（Monterey）向け互換版**を追加しました。

## ダウンロード

| パソコン | インストール用 | ZIP |
| --- | --- | --- |
| Windows 10／11 x64 | [EXE](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.2/MofuMouse-0.1.0-preview.2-win-x64.exe) | [ZIP](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.2/MofuMouse-0.1.0-preview.2-win-x64.zip) |
| macOS 13以降 · Apple Silicon | [DMG](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.2/MofuMouse-0.1.0-preview.2-mac-arm64.dmg) | [ZIP](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.2/MofuMouse-0.1.0-preview.2-mac-arm64.zip) |
| macOS 13以降 · Intel | [DMG](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.2/MofuMouse-0.1.0-preview.2-mac-x64.dmg) | [ZIP](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.2/MofuMouse-0.1.0-preview.2-mac-x64.zip) |
| macOS 12 · Apple Silicon | [互換版DMG](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.2/MofuMouse-0.1.0-preview.2-macos12-arm64.dmg) | [互換版ZIP](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.2/MofuMouse-0.1.0-preview.2-macos12-arm64.zip) |
| macOS 12 · Intel | [互換版DMG](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.2/MofuMouse-0.1.0-preview.2-macos12-x64.dmg) | [互換版ZIP](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.2/MofuMouse-0.1.0-preview.2-macos12-x64.zip) |

macOS 13以降は通常版、Montereyは名前に `macos12` がある互換版を選んでください。macOS 11以前、Windows 32bit／ARM版は今回の対象外です。別途Node.jsやPythonは不要です。

## 起動・更新

WindowsのZIPは全体を展開し、中の `MofuMouse.exe` を起動します。MacはDMGを開き、MofuMouseをアプリケーションへ移してください。動物アイコンのメニューから設定・一時停止・終了できます。

更新前に旧版を終了し、インストーラーを実行／ZIP全体を展開／Macアプリを置き換えます。preview.1の設定を引き継ぎます。自動更新はありません。[詳しい起動・更新手順](https://udteach.github.io/MofuMouse/download.html#start)

## 確認状況

Windowsでは実アプリの表示・個別選択・混在表示・4サイズ・一時停止／再開と同梱画像を確認しています。配布物は同一コミット・同一カタログの3,416枚のPNGを照合し、SHA-256を添付しています。

MacはApple Silicon／Intelの構成、最小OS設定、アドホック署名の整合性、同梱ソース・画像をCIで確認しています。**Mac実機での表示・操作、およびMonterey実機の起動は未確認です。**

Windowsコード署名、Apple Developer ID署名・公証は未実施です。Mac初回の警告は[起動手順](https://udteach.github.io/MofuMouse/download.html#mac-first-open)と[Apple公式の案内](https://support.apple.com/ja-jp/102445)を確認してください。

開発途中の先行版です。動物の寸法・毛色・歩行は引き続き調整します。ほかの動物の全色完成版ではなく、歩行コマ数は種類によって異なります。画像・動きの制作にはAIを使用しています。

[SHA256SUMS.txt](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.2/SHA256SUMS.txt) · [不具合を報告](https://github.com/UDteach/MofuMouse/issues) · [過去のリリース](https://github.com/UDteach/MofuMouse/releases)
