# MofuMouse Degu Edition 要件調査メモ

調査日: 2026-06-20
作業フォルダ: `D:\開発\MofuMouse`

## 現状

- `D:\開発\MofuMouse` は空の作業フォルダで、Git リポジトリではない。
- 既存実装がないため、まずは「チューチューマウスの現代版」兼「デグーのデスクトップマスコット」として要件を切る。
- 過去メモリには、ユーザーが Windows デスクトップ上でデグーがうろちょろするアプリ、タスクトレイメニュー、タイピング反応、複数の現実的な毛色、ImageGen-first の素材制作を希望した記録がある。

## 参照した主な情報

- 公式: チューチューマウス for WIN32
  - https://www.ikehouse.co.jp/onlinesoft/tyu32/
  - 1993 年公開。カーソルが「はい」「いいえ」などのボタンへ走り、元の場所へ戻ることで操作を楽にするソフト。
  - 特徴として、ボタン/フォーカスへの走行、マウス移動量削減、IE ナビ、マウス下ウィンドウのアクティブ化、画面端ワープなどが挙げられている。
- 窓の杜ライブラリ: チューチューマウス
  - https://forest.watch.impress.co.jp/library/software/tyumouse/
  - 新しいダイアログの OK ボタンなどへ自動移動。ハムスター型カーソルが鳴きながら走る。クイックスクロール、クイックキー、IE ナビも説明されている。
- 窓の杜回顧記事
  - https://forest.watch.impress.co.jp/docs/shseri/pastprize/1037222.html
  - 1998 年の窓の杜大賞受賞ソフト。ボール式マウス時代に移動量削減が実用価値だったこと、IE の戻る/進むをマウスボタン長押しで操作する機能が注目されたことが確認できる。
- Smithsonian National Zoo: Degu
  - https://nationalzoo.si.edu/animals/degu
  - デグーはアンデス斜面原産の小型で社会的なげっ歯類。茶色い背中、白い腹、ブラシ状の尾、地下トンネル、警戒音、日中活動、グルーミング/遊び/砂浴びが特徴。
- RSPCA: Keeping degus as pets / Understanding degu behaviour
  - https://www.rspca.org.uk/adviceandwelfare/pets/rodents/degus
  - https://www.rspca.org.uk/adviceandwelfare/pets/rodents/degus/behaviour
  - 小型・社会的・声でよくコミュニケーションする、日中活動、掘る/かじる、黄色い歯、砂浴び、隠れ場所、仲間、予測可能な明暗サイクルが重要。
- Animal Diversity Web: Octodon degus
  - https://animaldiversity.org/accounts/Octodon_degus/
  - 体重 170-300g、黄色がかった茶色の背中、クリーム色の腹、大きく黒っぽい耳、房状の尾。社会性、半地中性、朝夕に活発、音/匂い/視覚でコミュニケーション。
- Microsoft Learn: Windows 実装関連
  - SendInput: https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-sendinput
  - UI Automation Button: https://learn.microsoft.com/en-us/windows/win32/winauto/uiauto-supportbuttoncontroltype
  - Layered Windows: https://learn.microsoft.com/en-us/windows/win32/winmsg/window-features
  - Extended Window Styles: https://learn.microsoft.com/en-us/windows/win32/winmsg/extended-window-styles

## OSS/GitHub 追加調査

調査時点: 2026-06-20。スター数・更新日は GitHub API で確認。

| 種別 | プロジェクト | 状況 | 持ち帰る設計 |
| --- | --- | --- | --- |
| mouse jiggler | `arkane-systems/mousejiggler` | C#、約 1.4k stars、2026-06-18 更新。README では Windows に fake mouse input を入れてスクリーンセーバー等を避ける用途、Normal/Zen/Circle/Linear、マウス操作中の自動 pause、ランダム間隔、tray、CLI を説明。ライセンスは GitHub API 上は `NOASSERTION` なのでコード流用は避ける。 | 「物理ジグル」と「Zen/仮想ジグル」を分ける。ユーザーがマウスを動かしたら止める。隠蔽機能は入れない。 |
| mouse activity simulator | `sw3103/movemouse` | C#、GPL-3.0、約 772 stars、2026-06-18 更新。Windows ロック/スリープ防止のための user activity simulation。 | GPL なのでコード流用は避け、シナリオ設計だけ参考にする。時間帯/条件付き実行の思想は有用。 |
| mouse mover | `ebellocchia/auto_mouse_mover` | C#、MIT、約 43 stars、2026-02-06 更新。カーソル位置が前回から変わっていない時だけ動かし、戻すことで同じ場所に留める。間隔と移動ピクセル数を設定可能。 | 1px 移動機能の MVP 参考。ユーザー操作中は邪魔しない、往復で元位置へ戻す、画面外へ出さない。左クリック機能は入れない。 |
| keep-awake utility | Microsoft PowerToys Awake | PowerToys の一部、MIT、約 135k stars、2026-06-20 更新。Microsoft Learn では、電源設定を変更せず PC を起こし続け、無期限/時間指定/期限日時、display-on、tray、CLI、PID 連動を提供。ロック画面では機能しない制約も明記。 | OS 公式の keep awake は「画面/スリープ防止」として別モードにする。終了時に必ず解除する。ロック画面や共有環境の注意を表示する。 |
| desktop pet overlay | `SeakMengs/WindowPet` | Tauri + React、MIT、約 628 stars、2026-06-20 更新。click through、auto startup、pet above taskbar、settings、custom pet など。 | カーソル横アシストでも、透明 overlay、click-through、taskbar 上表示、設定画面、カスタム素材の考え方を参考にする。 |
| desktop mascot | `DalekCraft2/Shimeji-Desktop` / Shimeji-ee 系 | Java、約 38 stars、2026-06-19 更新。XML で行動とアニメーションを定義する desktop mascot。ライセンスは API 上 `NOASSERTION`。 | コード流用は避ける。アクションをデータ駆動にして、素材と動きを差し替えやすくする発想だけ採用。 |
| UI Automation helper | `yinkaisheng/Python-UIAutomation-for-Windows` | Python、Apache-2.0、約 3.5k stars、2026-06-17 更新。MFC/WinForms/WPF/Qt/Chrome/Electron 等の UI Automation 対応アプリを探索できる。 | Python 試作なら有力。C# 実装でも「UIA で候補ボタンを探索し、名前/矩形/パターンを見る」検証に使える。 |

関連する Windows 既存機能:

- Windows/.NET には `SnapToDefaultButton` があり、既定ボタンへポインターを移す設定が存在する。ただし Raymond Chen の解説では、標準 dialog manager を使わないカスタムダイアログでは効かない。MofuMouse はこの穴を UI Automation で補完できる。
- `SetThreadExecutionState` はスリープやディスプレイ電源オフの抑止に使えるが、Microsoft Learn では「スクリーンセーバーは止めない」と明記されている。スクリーンセーバー防止には 1px 移動/仮想入力系が必要。
- `PowerSetRequest` は `PowerRequestDisplayRequired` と `PowerRequestSystemRequired` を分けて扱える。ディスプレイを保つには system request も併用する注意がある。
- `GetLastInputInfo` は session-specific な最後の入力時刻を取れる。1px 移動やアシスト表示は、ユーザーが実際に操作中かどうかの判定に使う。

## 元ネタから引き継ぐべきコア価値

1. カーソル横でアシストする実用性
   - 新規ダイアログ、確認ダイアログ、フォーカス対象のボタンなどへカーソルを移動する。
   - ただし現代版では安全上、自動クリックは MVP から外す。
   - ユーザー希望により、MVP は「カーソルの横にいるデグーが操作を助ける」体験を主役にする。

2. 「走って向かう」キャラクター性
   - ただ瞬間移動するのではなく、デグーが走る/鳴く/戻ることで楽しくする。
   - 元ネタのハムスター相当をデグーに置き換えるだけでなく、砂浴び・掘る・かじる・警戒・グルーミングを入れる。

3. 常駐ユーティリティ
   - タスクトレイ常駐、オン/オフ、設定、終了、起動時開始が必要。

4. 古い機能の取捨選択
   - 採用候補: ボタンへの移動、元位置へ戻る、クイックスクロール風の補助、画面端ワープ。
   - 見送り候補: パスワード表示、ブラウザ固有の IE/Netscape ナビ、危険な自動決定。

## デグーらしさの要件

### 見た目

- 基本形は「大きめの耳、ずんぐりした胴体、短めの脚、細長い尾とブラシ状の尾先」。
- 標準カラーはアグーチ/黄褐色系。腹は白またはクリーム色。
- バリエーション候補:
  - natural agouti / brown
  - blue / gray
  - black / dark brown
  - sand / cream
  - white / high pied
  - pied / spotted
- 尾先の房、耳、目の周りの淡い色、黄色い歯は小さいサイズでも識別できるようにする。

### 動き

- 必須アニメーション:
  - walk/run
  - idle blink
  - sniff
  - groom
  - chew/nibble
  - dig
  - dust bath roll
  - freeze/alarm
  - sleep
- カーソル補助時:
  - ターゲットへ小走り。
  - 鳴き声は現在の仕様に含めない。
  - 到着後に鼻先でボタン付近を探る、または自然なしぐさで知らせる。
  - 設定により元の場所へ戻る。
- アンビエント時:
  - タスクバー近くや画面端をうろちょろする。
  - タイピング量に反応して覗く、走る、毛づくろいする。
  - 日中は活発、夜は寝やすいスケジュールにできる。

### 行動ロジック

- 社会性:
  - MVP は 1 匹でもよいが、将来的に 2 匹モードやペア行動を入れられる設計にする。
- 掘る/隠れる:
  - 画面端、タスクバー、ウィンドウ裏へ潜るような演出。
- かじる:
  - 実際の UI を壊さず、ウィンドウ枠をかじるふりだけにする。
- 砂浴び:
  - 設定画面または待機中に一定確率で発生。
- 警戒:
  - 急なマウス移動、大きなウィンドウ変化、通知の出現などで freeze/alarm。
- 尾:
  - ドラッグ操作は胴体を掴む表現にし、尾を引っ張る演出は避ける。

## 機能要件案

### MVP

- Windows 10/11 x64 のデスクトップ常駐アプリ。
- 透明・常に手前・クリック透過のデグー overlay。
- デグーは通常、カーソルの横または少し後ろにいる。
  - カーソルに密着しすぎず、クリック対象を隠さない。
  - 入力欄やボタンの上に長居しない。
  - ユーザーが大きくマウスを動かした時は追従を遅らせ、邪魔をしない。
- タスクトレイメニュー:
  - 有効/一時停止
  - Windows 起動時に始めるオン/オフ
  - カーソル補助オン/オフ
  - カーソル横デグー表示オン/オフ
  - Keep Awake オン/オフ
  - Keep Awake 時間指定
  - 毛色変更
  - 設定
  - 終了
- ダイアログ/前面ウィンドウ検出:
  - 現在の MVP では classic Win32 の子 `Button` 走査と x64 UI Automation 走査を実装し、OK/保存/適用などの安全候補を sniff 表示に使う。
  - UI Automation fallback は前面アプリが複雑でもMofuMouseを固めないよう、短いタイムアウトと単一起動ガードを持つ。
  - classic Win32 の子 `Button` が見つかる場合はそれを先に使う。classic が空の時だけ UI Automation にフォールバックし、UIA の遅延で classic ダイアログの物理移動が詰まるのを避ける。
  - UI Automation では Button control の `BoundingRectangle` / `ClickablePoint` / `Name` を読む。x86 は classic Win32 へフォールバックする。
  - OK/Yes/No/Cancel/Apply などを候補にする。
  - 既定ボタンまたはユーザー設定の優先順位でターゲットを選ぶ。
- カーソル移動:
  - 物理カーソルを動かすか、デグー overlay がターゲットを示すだけにするかを設定化。
  - 物理移動はユーザーがマウスを動かしたら即キャンセル。
  - 自動クリックは実装しない。
- カーソル横アシスト:
  - ボタン候補が見つかった時、デグーがカーソル横から対象方向を向く。
  - MVP の実装済み動作は sniff assist と、明示オン時だけの物理カーソル移動/元位置復帰。自動クリックは行わない。
  - 危険度の低い候補は「ここだよ」表示。危険候補は強調せず、必要なら注意だけ出す。
  - 見失い対策として、短い「しっぽ/足跡」トレイルや現在位置ハイライトを任意で出せる。
- 入力反応:
  - キー内容は読まず、打鍵イベント/頻度だけを見る。
  - 明示的にオンにした場合のみグローバル入力監視を使う。
- 1px 移動 / スクリーンセーバー防止:
  - MVP に含める。ただし初期値はオフ。
  - 「最小物理ジグル」: 一定時間入力がない時だけ、カーソルを 1px 動かしてすぐ戻す。
  - 「仮想ジグル」: 可能ならポインター表示位置を変えず、入力アイドルだけをリセットする。
  - 「OS Keep Awake」: `SetThreadExecutionState` / `PowerSetRequest` 系でスリープとディスプレイ電源オフを抑止する。スクリーンセーバー防止とは別扱いにする。
  - いずれも時間指定を必須にできる設定を用意する。
- 設定保存:
  - JSON などのローカル設定。
  - 初期値は安全側: 自動クリックなし、危険ボタン回避。
- 自動起動:
  - 既定はオフ。
  - ユーザー権限の `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` に MofuMouse の実行ファイルだけを登録する。
  - オフにしたら Run 登録を削除する。

### v1.1 以降

- 元位置へ戻る/戻らない、移動速度、ターゲットボタン優先順位の詳細設定。
- クイックスクロール風の補助。
- 画面端ワープ。
- 複数デグー表示。
- デグーの部屋/巣穴/砂場などのミニ演出。
- アニメーションパック差し替え。
- 作業モード:
  - 会議/発表モード: 30分、1時間、2時間などの Keep Awake プリセット。
  - 期限日時モード: 指定時刻まで Keep Awake。
  - PID 連動モード: 指定プロセスが生きている間だけ Keep Awake。
  - 低バッテリー時の自動停止/警告。
- アシスト学習:
  - よく選ぶボタン種別の優先順位をローカルに保存する。
  - 危険操作は学習対象から除外する。

### 見送るべき機能

- パスワード表示。
- UAC/セキュアデスクトップ操作。
- 管理者権限アプリへの入力注入を前提にした動作。
- 自動クリックによる Yes/Delete/Format/送信などの確定。
- キー入力内容の収集、送信、ログ保存。
- 監視回避や隠蔽を目的にした機能。
- tray icon を完全に隠して Task Manager 以外で止められない挙動。
- Keep Awake の無期限オンを初期値にすること。
- 左クリック付き mouse jiggler。

## 追加機能案: カーソル横アシスト

ユーザー希望により、MVP の主役は「デグーがカーソルの横で助けてくれる」ことに固定する。

### アシストの基本状態

- `follow`: カーソルから 24-48px 程度離れて付いてくる。
- `peek`: ダイアログや候補ボタンを見つけたら、対象方向を向く。
- `sniff`: 安全な候補ボタン付近を鼻先で探る。人間的な指差しは使わない。
- `run`: ユーザーが明示した時だけ、カーソルまたはデグーが対象近くへ移動する。
- `guard`: 長時間入力がない時、Keep Awake の状態を示す。
- `sleepy`: Keep Awake が切れる直前やバッテリー低下時に眠そうな表示で知らせる。

### ボタン補助の優先順位

1. Enter で実行される既定ボタン。
2. OK / Apply / Save など低リスクの確定。
3. Cancel / Close / No など取り消し系。
4. Delete / Remove / Format / Send / Purchase / Pay / Sign / Publish など危険語を含む候補は自動移動しない。

### Windows Snap To との関係

- Windows の `SnapToDefaultButton` がオンなら、それを尊重する。
- 標準 Snap To が動いた場合、MofuMouse は追従して横に出るだけにする。
- Snap To が効かないカスタムUIでは、MofuMouse が UI Automation で候補を検出して補完する。
- 移動前には foreground window を再確認する。背景ウィンドウへ勝手に移動しない。

## 追加機能案: 1px 移動 / Keep Awake

### モード

1. `Off`
   - 何もしない。初期値。

2. `OS Keep Awake`
   - `SetThreadExecutionState(ES_CONTINUOUS | ES_SYSTEM_REQUIRED)` を基本にする。
   - 表示維持が必要な時だけ `ES_DISPLAY_REQUIRED` または `PowerRequestDisplayRequired` を使う。
   - スクリーンセーバーは止めないことを UI に明記する。

3. `1px Jiggle`
   - `GetLastInputInfo` で一定時間の無操作を確認する。
   - カーソル位置が前回から変わっていない時だけ `(x+1, y)` へ移動し、短時間で `(x, y)` へ戻す。
   - 画面端では内側方向へ動かす。
   - ドラッグ中、ゲーム/全画面中、リモート操作中、ユーザーが直近に動かした時は止める。

4. `Virtual Jiggle`
   - ポインター表示位置を変えずに idle reset だけを狙う。
   - 一部アプリ独自の idle detection には効かない可能性を明記する。
   - 物理ジグルより優先候補だが、検証で効かなければ experimental 扱い。

5. `Presentation / Demo`
   - 15分/30分/1時間/2時間/指定時刻まで。
   - 終了前にデグーが眠そうに通知する。
   - 終了時は必ず power request を解除し、通常状態へ戻す。

### 安全ルール

- 企業監視回避をうたわない。用途は発表、長時間処理、作業中の不要なスクリーンセーバー防止に限定して説明する。
- 停止手段は常に tray とホットキーで見える場所に置く。
- 無期限モードは警告付き。初回は時間指定を推奨する。
- ロック画面・UAC・別セッションでは期待通りに動かないことを明記する。

## 技術要件

### Windows UI/入力

- ターゲット検出は UI Automation を第一候補にする。
  - Microsoft Learn では Button control に `BoundingRectangle` や `ClickablePoint` が定義されている。
- 物理カーソル移動は `SetCursorPos` または `SendInput` 相当。
  - `SendInput` は UIPI の制約を受け、同等以下の整合性レベルのアプリにしか入力注入できないため、管理者権限アプリでは失敗しうる。
- overlay は layered window を使う。
  - per-pixel alpha、常時手前、クリック透過が必要。
  - `WS_EX_TOPMOST`, `WS_EX_NOACTIVATE`, `WS_EX_LAYERED` などが候補。

### 候補スタック

1. C# / WPF or WinUI
   - 長所: Windows API/UI Automation/高 DPI/配布との相性がよい。
   - 短所: アニメーション/素材パイプラインを少し設計する必要がある。
   - カーソル横アシストと 1px/Keep Awake を主役にするなら第一候補。

2. Python / PySide6
   - 長所: 実装が速い。Pillow/pystray/pywinauto/comtypes/PyInstaller と組み合わせやすい。
   - 短所: 透明 overlay、グローバル入力、DPI、exe 配布の詰めが必要。
   - プロトタイプ重視なら第一候補。

3. Tauri / Electron
   - 長所: アニメーション UI を作りやすい。
   - 短所: OS 入力/UI Automation 連携が重くなりやすい。Electron は常駐ペットにしては重い。
   - デザイン優先なら検討。

## 非機能要件

- 安全性:
  - 自動クリックなし。
  - 危険そうなボタンは回避または確認。
  - ユーザー操作が始まったら移動キャンセル。
  - 緊急停止ホットキーとトレイ一時停止を用意。
  - 1px 移動と Keep Awake は常に状態表示する。
  - Keep Awake は終了時・クラッシュ時に解除される設計にする。
- プライバシー:
  - キー内容、ウィンドウ内容、ファイル名などを外部送信しない。
  - 入力ログを残さない。
- パフォーマンス:
  - アイドル CPU はほぼ 0% を目標。
  - アニメーション更新は必要時だけ。
  - メモリは軽量常駐として過大にしない。
- UX:
  - 邪魔ならすぐ止められる。
  - 動き/音/表示サイズを調整できる。
  - Reduced motion / mute / 低刺激モードを用意。
- 互換性:
  - Windows 10/11。
  - マルチモニター、高 DPI、タスクバー位置変更に対応。
  - 管理者権限ウィンドウや UAC では動かない場合があることを明示。

## 素材要件

- ImageGen-first で制作。
- 透過 PNG または sprite sheet。
- 最初に必要な素材:
  - 6 毛色 x idle/run/groom/dig/dust-bath/freeze/sleep の最低セット。
  - トレイアイコン。
  - 小サイズでも見える cursor buddy 用縮小版。
- 1 アクションあたり 6-12 フレームを目安にする。
- 小型表示で潰れないよう、リアル寄りでも輪郭と耳/尾先は強調する。

## QA/検証要件

- UI Automation / classic scanner のターゲット統合と重複除去のユニットテスト。
- ダミーダイアログでの統合テスト:
  - OK/Cancel
  - Yes/No
  - Apply
  - disabled button
  - owner window付き modal dialog
- 物理カーソル移動の安全テスト:
  - ユーザー入力でキャンセルされる。
  - ターゲット消失時に止まる。
  - 画面外へ行かない。
  - 独立した Win32 fixture で、OK は選ばれ、Delete は危険候補として避けられ、disabled Apply は選ばれない。
  - 飼育手帳ルールで OK を block し Later を prefer した場合、Later へ移動する。
  - 実アプリQAとして Notepad の未保存状態を使い、`保存しない` / Don't Save / Discard 系の候補が dangerous 扱いになり safe list に入らないことを確認する。
  - 実アプリQAとして Calculator を使い、`ApplicationFrameHost.exe` 配下の modern Windows app から安全候補を取得でき、safe list に dangerous 候補が混ざらないことを確認する。
- カーソル横アシストの QA:
  - カーソル、ボタン、入力欄を隠さない。
  - 高 DPI でカーソルとの距離が破綻しない。
  - マルチモニター境界で表示が消えない。
  - `scripts/verify_display_qa.ps1` で x86/x64 main app の DPI awareness と、実モニター矩形の union が Windows 仮想スクリーン座標と一致することを確認する。
  - 全画面アプリやゲームでは自動的に控える。
- Keep Awake QA:
  - OS Keep Awake が `SetThreadExecutionState` / `PowerSetRequest` の失敗を検出して表示する。
  - `PowerClearRequest` 相当の解除が終了時に必ず行われる。
  - 1px Jiggle が元位置へ戻る。
  - ユーザーがマウスを動かしたら jiggle しない。
  - 画面端でも画面外に出ない。
  - スクリーンセーバー防止の効き方を OS Keep Awake と 1px Jiggle で分けて手動確認する。
- マルチモニター/DPI の手動 QA。自動QAは DPI awareness と仮想スクリーン union まで確認する。物理マルチモニター機材での画面端実走確認は release gate として残す。
- トレイ常駐/終了/一時停止/自動起動の smoke test。
- パッケージ後 exe smoke test。
- 視覚 QA:
  - 透明部分がクリックを奪わない。
  - Alt+Tab やタスクバーに不要表示されない。
  - デグーがボタンや入力欄を長時間隠さない。

## 推奨 MVP 方針

最初は「カーソル横でアシストしてくれるデグー」にする。

- 中核はダイアログのボタン検出、カーソル横追従、対象ボタンの示唆、必要時だけ安全に移動する挙動。
- 自動クリックは入れない。
- 1px 移動/スクリーンセーバー防止は MVP に含める。ただし初期値はオフで、時間指定を推奨する。
- デスクトップを自由にうろちょろする要素は弱め、基本はカーソルの相棒として振る舞う。
- 毛色変更、控えめなタイピング反応、短い砂浴び/毛づくろい演出は MVP に含めるが、複数デグーや高度な学習は後回し。
- 旧チューチューマウスのパスワード表示や IE 固有ナビは現代版から外す。

この方針なら、元ネタの価値を保ちつつ、デグー版としての楽しさ、スクリーンセーバー防止の実用性、安全性を両立しやすい。

## 未決事項

1. 物理カーソル移動を初期オンにするか、最初はデグーが安全候補を sniff / 探索で知らせるだけにするか。
2. 1px Jiggle を MVP で正式機能にするか、experimental にするか。
3. 技術スタックを C# 系にするか、Python/PySide6 で先に試作するか。
4. 素材テイストをリアル寄り、ドット絵寄り、ぬいぐるみ寄りのどれにするか。
5. アプリ名を `MofuMouse` のままにするか、デグー専用名を付けるか。
## 2026-06-20 追加要件

- 既定表示サイズは約 32px とし、ユーザー設定 `SpriteSizePx` で調整できる。
- overlay で pet name を常時表示しない。名前は設定・トレイ・将来の詳細画面で扱い、画面上ラベルは `ShowPetName` が true の場合だけ表示する。
- cursor buddy はカーソルの真横に置く。既定オフセットは `4,-22`。
- 右端では cursor 左側へ反転し、仮想スクリーン内に clamp する。
- カーソル移動中は静止アクションより movement animation を優先し、追従位置を補間して滑らかに近づける。
- `CoatColor` 設定とトレイメニューで、`agouti` / `gray` / `dark` / `cream` / `white` / `pied` を選べる。
- 全 6 毛色に `idle` / `walk` / `run` を用意する。
- Keep Awake または 1px Jiggle 有効時は、静止中に `guard` ポーズを表示できる。
- 無操作アクションには `nibble` / `groom` / `patpat` / `sniff` / `sleepy` / `dig` / `roll` を含める。
- 毛色別pose素材が未生成の場合は、agoutiの別poseへ色飛びさせず、同じ毛色のidleへフォールバックする。
- 正式アセットは必ず ImageGen で 1 枚ずつ生成する。sprite sheet / reference sheet からの切り抜きは、途切れや画風差のQAを通るまで runtime 採用しない。
- runtime 採用QAは、全身が切れていないこと、透明背景、32px可読性、既存runtime素材との画風一貫性を必須とする。

## 2026-06-21 追加要件: Chuchu Mouse を超えるための再抽出

### Product principle

MofuMouse は「勝手に押すツール」ではなく、「カーソルのすぐ横で、次に何ができるかをデグーが安全に案内するアシスタント」とする。

- 初期値では自動クリックしない。
- 危険な操作は候補表示や自動移動の対象から外す。
- ユーザーがマウスを動かしたら、補助動作は即時キャンセルする。
- 初心者向けラベルを先に出し、詳細設定は Advanced に分ける。

### Feature categories

| Category | Requirement | Notes |
| --- | --- | --- |
| Cursor companion | The degu stays beside the cursor, flips based on movement direction, avoids covering buttons, and uses walk/run while moving. | Current MVP has the base overlay. Walk/run assets need ongoing QA because their foreground is lower-profile than idle. |
| Target assist | Detect safe dialog buttons with UI Automation and let the degu notice them with a sniff/investigate pose. | OK/Apply/Save first. Cancel/No can be optional. Dangerous labels are denied by default. |
| Safe movement | Optional cursor movement to a safe target, then optional return to origin. | Disabled by default until target scoring and cancellation QA are solid. |
| Per-app rules | Save app/window/button preferences locally. | Implemented as local `TargetRules`. Never store window text contents beyond minimal labels needed for user-authored rules. |
| Keep awake | OS keep-awake, 1px physical jiggle, schedules, presentation presets, and visible status. | Explain that OS keep-awake and screensaver prevention are different. |
| Quick actions | Browser back/forward, show desktop, start menu, quick key palette, quick scroll. | Browser/show desktop/start are tray actions. Quick Key is implemented with a fixed safe allowlist. Quick Scroll is implemented as fixed safe tray up/down actions with a configurable line count. |
| Launch at login | Start MofuMouse after Windows sign-in. | Implemented via HKCU Run key, off by default, tray/settings controlled. |
| Degu behavior | Idle, walk, run, return, guard, sniff, groom, nibble, patpat, dig, dust bath roll, sleepy, alert/freeze. | Official runtime art must be ImageGen one-by-one, not sheet cutouts. Human-like point poses are rejected. |
| Appearance | Name, coat color, size, opacity, reduced motion, preview. | Name is not always shown on overlay by default. Sound is not in current scope. |
| Updates | GitHub Releases check, x86/x64 asset choice, local download cache, prerelease opt-in, rollback note. | Asset selection/download, CLI install/rollback, and tray-driven confirmation/restart UX are implemented. Public release remains blocked until all release gates pass. |
| Privacy and safety | Local-only settings, no password capture, no keystroke content logs, no screenshots unless explicit QA capture. | No UAC/secure desktop automation. |

Current settings implementation:

- Wails v2 + React + Fluent UI control center exists for x64 as `mofumouse-control-x64.exe`.
- The control center reads/writes the same local `core.Settings` JSON model.
- Sound controls are intentionally not exposed.
- The Win32 settings window remains available through `--legacy-settings` and as fallback.
- x86 keeps the Win32 settings path until Wails/x86 packaging is explicitly solved.
| Accessibility | Keyboard-friendly settings, clear labels, high contrast/reduced motion options, screen-reader names. | Modern settings UI must support tab navigation. |
| QA | Unit tests, x86/x64 builds, console-free subsystem check, overlay screenshots, multi-monitor/DPI checks, UIA sample dialogs. | These gates block release. |

### Modern settings requirements

The settings UI must move from a flat Win32 form to a tabbed control center. Required tabs:

1. Home: current status, quick toggles, update state.
2. Companion: cursor-side placement, movement, idle actions, reduced motion.
3. Assist: button detection, safe target rules, per-app behavior.
4. Keep Awake: OS keep-awake, 1px jiggle, schedule, presentation presets.
5. Appearance: name, size, coat, opacity, asset preview.
6. Updates: version, GitHub Releases check, x86/x64 asset selection.
7. Advanced: JSON path, diagnostics, QA capture, reset/export/import.

### Explicitly out of scope

- Automatic click on Yes/Delete/Format/Pay/Send/Publish or similar risky actions.
- Password reveal, password text capture, or key content logging.
- UAC / secure desktop automation.
- Stealth or hidden keep-awake behavior with no visible tray/status control.
- Releasing GitHub/Pages artifacts before requirements, assets, and QA gates pass.

## 2026-06-21 追加要件: high-resolution asset remake

Current 32px runtime sprites become jaggy when the user increases `SpriteSizePx`. The production asset pipeline must be rebuilt around higher-resolution masters.

- Regenerate the official runtime degu set with ImageGen, one pose at a time.
- Keep the ImageGen source/master large enough for future exports. Treat 64px as the minimum runtime tier, not the only source.
- Export multiple runtime tiers: `32`, `48`, `64`, and `96` px.
- The overlay renderer should choose the nearest larger tier and downscale when possible. It should avoid upscaling a 32px sprite.
- Runtime QA must compare 32/48/64px display for every accepted pose.
- Walk/run must be redrawn as cursor-buddy poses with a body height closer to idle, not just scaled horizontally.
- All official runtime poses should be regenerated for visual consistency, because mixing old 32px assets and new 64px assets will make the app look uneven.
- `return` pose is a dedicated ImageGen runtime asset for the optional TargetReturn motion. It must be visible only as a short transient action and must not replace same-coat idle fallback for non-agouti coats until matching color variants exist.
- Until the multi-tier renderer is implemented, large display sizes are considered preview quality.

## 2026-06-21 additional requirement status: safe cursor movement

- `TargetMove` controls physical cursor movement to a safe detected button. It is off by default.
- `TargetReturn` controls returning to the previous cursor position. It is off by default.
- Movement only uses `SafeForegroundAssistTargetsWithRules`, so dangerous labels and user-authored block rules remain excluded.
- Movement does not click.
- Return is canceled if the user moves the mouse before the return delay.

## 2026-06-21 additional requirement status: 飼育手帳 / per-app assist rules

- `TargetRules` stores local user-authored rules in settings JSON.
- A rule can match by foreground app executable name, window title, and button label. Each non-empty field is a case-insensitive partial match.
- Rule actions are `prefer` and `block`.
- `prefer` can move a safe candidate earlier, but cannot make a built-in dangerous target safe.
- `block` marks a matching candidate as dangerous for sniff assist and physical movement.
- Control Center Assist tab provides a beginner-facing rule editor.
- QA evidence:
  - `scripts/verify_target_move_qa.ps1 -ConfigPath .codex/qa/qa-target-move-settings.json`
  - `scripts/verify_target_move_qa.ps1 -ConfigPath .codex/qa/qa-target-move-return-settings.json -ExpectReturn`
  - `scripts/verify_target_move_qa.ps1 -ConfigPath .codex/qa/qa-target-move-return-settings.json -ExpectReturn -SimulateUserMove`
  - `scripts/verify_notepad_assist_qa.ps1`
