写真風の小動物がマウスカーソルについてくるMofuMouseに、「のんびりモード」を追加しました。

[**Webでおためし**](https://udteach.github.io/MofuMouse/try/) · [**ダウンロードと使い方**](https://udteach.github.io/MofuMouse/download.html) · [公式サイト](https://udteach.github.io/MofuMouse/)

![MofuMouseの10種の表示例](https://udteach.github.io/MofuMouse/assets/mofumouse-preview-ten-animals.png)

## 今回の更新

- **のんびりモード**を追加。通知領域／メニューバーの「追いかけ方 → のんびり」で切り替えます。ゆっくり動き出し、加速・減速しながら追いかけ、カーソルが止まると追いついて休みます。
- 移動の最高速度を制限し、高速にカーソルを動かしても動物が飛ぶように移動しないようにしました。細かな揺れで向きが頻繁に変わるのも抑えます。広い画面では到着まで時間がかかります。
- 古い移動先を溜めず、現在のカーソル位置を追いかけます。歩行アニメは動物自身の移動に合わせて再生します。
- 通常モードも選べます。追いかけ方の設定は次回起動へ引き継ぎ、以前の設定では通常モードから始まります。
- **Webデモでも通常／のんびりを切り替え**できます。

収録内容はpreview.2と同じ10種・26種類、デグー全10色。1〜10匹・32／48／64／96pxに対応します。今回、動物画像は変更していません。Windows、Mac通常版、macOS 12向け互換版を提供します。

## ダウンロード

| パソコン | インストール用 | ZIP |
| --- | --- | --- |
| Windows 10／11 x64 | [EXE](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.3/MofuMouse-0.1.0-preview.3-win-x64.exe) | [ZIP](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.3/MofuMouse-0.1.0-preview.3-win-x64.zip) |
| macOS 13以降 · Apple Silicon | [DMG](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.3/MofuMouse-0.1.0-preview.3-mac-arm64.dmg) | [ZIP](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.3/MofuMouse-0.1.0-preview.3-mac-arm64.zip) |
| macOS 13以降 · Intel | [DMG](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.3/MofuMouse-0.1.0-preview.3-mac-x64.dmg) | [ZIP](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.3/MofuMouse-0.1.0-preview.3-mac-x64.zip) |
| macOS 12 · Apple Silicon | [互換版DMG](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.3/MofuMouse-0.1.0-preview.3-macos12-arm64.dmg) | [互換版ZIP](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.3/MofuMouse-0.1.0-preview.3-macos12-arm64.zip) |
| macOS 12 · Intel | [互換版DMG](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.3/MofuMouse-0.1.0-preview.3-macos12-x64.dmg) | [互換版ZIP](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.3/MofuMouse-0.1.0-preview.3-macos12-x64.zip) |

macOS 13以降は通常版、Montereyは名前に `macos12` がある互換版を選んでください。macOS 11以前、Windows 32bit／ARM版は今回の対象外です。別途Node.jsやPythonは不要です。

## 起動・更新

WindowsのZIPは全体を展開し、中の `MofuMouse.exe` を起動します。MacはDMGを開き、MofuMouseをアプリケーションへ移してください。動物アイコンのメニューから設定・一時停止・終了できます。

更新前に旧版を終了し、インストーラーを実行／ZIP全体を展開／Macアプリを置き換えます。preview.1の設定を引き継ぎます。自動更新はありません。[詳しい起動・更新手順](https://udteach.github.io/MofuMouse/download.html#start)

## 確認状況

38件の自動テストで、速度・加速度の上限、停止後の到着、20分相当の連続入力、停止再開・スリープ復帰、設定保存を確認しました。Windowsでは通常／のんびりの実描画、4サイズ・10匹・クリック透過を確認しています。Web版はPC幅と390px幅で操作・表示を確認しました。実際の4K・高DPIマウス環境での速度感は未確認です。全配布物で同じカタログの3,416枚のPNGを照合し、SHA-256とビルド情報を添付しています。

MacはApple Silicon／Intelの構成、最小OS設定、アドホック署名の整合性、同梱ソース・画像をCIで確認しています。**Mac実機での表示・操作、およびMonterey実機の起動は未確認です。**

Windowsコード署名、Apple Developer ID署名・公証は未実施です。Mac初回の警告は[起動手順](https://udteach.github.io/MofuMouse/download.html#mac-first-open)と[Apple公式の案内](https://support.apple.com/ja-jp/102445)を確認してください。

開発途中の先行版です。動物の寸法・毛色・歩行は引き続き調整します。ほかの動物の全色完成版ではなく、歩行コマ数は種類によって異なります。画像・動きの制作にはAIを使用しています。

[SHA256SUMS.txt](https://github.com/UDteach/MofuMouse/releases/download/v0.1.0-preview.3/SHA256SUMS.txt) · [不具合を報告](https://github.com/UDteach/MofuMouse/issues) · [過去のリリース](https://github.com/UDteach/MofuMouse/releases)
