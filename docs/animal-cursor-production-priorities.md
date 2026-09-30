# 動物カーソル制作の優先順位と分担

## 現行の指定 — 2026-09-29

**最新ユーザー指定:** 現在の他種制作を保存・QAの区切りまで進め、その後はデグーの毛色・白斑10外見を先行する。実行入口は `.codex/tasks/degu-colors-priority-20260929.md`。下記の並行順序・42外見はこの指定前の全体計画/snapshotであり、新規生成の優先はデグーを上位にする。元の全種Goalは維持する。

ここを入口にし、下の「初期計画の履歴」を現在の生成指示として実行しない。

生成方式の最新更新（2026-09-29）: ユーザーが添付AGENTSを明示的に置き換え、個別poseを既定としつつ既存poseの2×2編集試験を許可した。Local Skill/Workflow5Aも同期済み。新規試験は現在の親指示の種・motion・4pose・call上限に限定する。過去の消費済みcapや方法評価から追加生成を推測しない。8pose以上の生成は許可されていない。

- 現在の対象は10種。既存19姿のElectron版には最大10匹の個別選択、保存、サイズ変更、速度追従がある。旧Goアグーチ専用版は現行基準ではない。
- 新しい歩行・色違い歩行は **30実ポーズの受入**へ移行中。正本は `assets/manifest/animal-cursor-walk30-handoffs.json` と各 `full30-v1/reference.json` / `timebase.json`。旧8pose handoffの `ready` だけでは新規walk生成を始めない。基準が不足する種は親の新Flow sourceを待ち、idleや出典・mapping準備を進める。
- 原動画の実fpsと連続コマを保持する。30枚にするための重複、補間、並べ替え、足の手描きは禁止。30枚に足位相が合わない候補は棄却する。デグーの2.25倍/556msはデグーだけの承認値で、新種へ転用しない。
- ImageGenは元poseの直接編集。通常は1pose/画像。後続ユーザー依頼の4pose/画像はWorkflow段階5Aの2×2 pilotとして、親が指定した種・色・pose ID・回数で試す。4pose成功を30pose完成に数えず、各シートの形・寸法・色・透過と同色walk↔idleを比較する。
- 既存の採用画像/動画/ハッシュは保持し、新revisionで候補を作る。候補、方法QA、source受入、ローカルruntime、公開版、物理Mac確認を別々に記録する。
- 先行版 `v0.1.0-preview.1` とページは公開済み。Actionsは許可された **1/1回を使用済み**。追加run/rerun/公開はこの制作指示から推定しない。以後の素材統合とWindows検証はローカルで行う。

最新の進行入口は `.codex/tasks/mofumouse-current-20260929-status-and-live-lanes.md`。`.codex/tasks/walk30-source-upgrades.md` は基準種の昇格履歴として保持する。現在のruntime/sourceはそれぞれ `electron-prototype/app/media/manifest.json` と `assets/manifest/animal-cursor-ten-species.json` を検証する。履歴内のプロセスID、枚数や個別指示はその時点の記録。

確認済み色の範囲は `.codex/qa/confirmed-coat-scope-20260929/scope.json` と `aliases-review/REPORT.md` を照合した **42外見・84組のwalk/idle**。基本10色は別枠であり、42は完成数ではない。内訳はchinchilla9、hamster12、guinea9、rabbit9、degu2、ferret1。旧40組の台帳だけでは全対象を網羅しない。未確認色、別品種・毛種、保留中の名称を自動追加しない。暗眼/ピンク眼のwhite guineaは別外見、hamster banded/patchworkは顔の白模様を固定差分として維持する。Rabbit blue-eyed-whiteの旧除外記録は親判断待ちで、この42に含めない。

### 現行の所有境界

- 親: 基準動物Flow、共通source/handoff/registry、Electron統合、ローカルWindows、Mac source kit、公開版の固定記録。
- 既存担当 `01a0e2c7-ff04-7db3-bd76-7286a1a2d631`: degu等の確認済み色。親からの最新の具体的なbatch指示と自身の台帳を使用。guinea/rabbit/macaroni専任laneと重複しない。旧40組Goalを再開しても新walk30基準を下げない。
- guinea専任 `01a0e83b-868e-7b90-92f7-c2bb0d602277`、rabbit専任 `01a0e83b-9160-77c3-aafe-7e538405ba98`、macaroni専任 `01a0e83b-9e19-7dd1-ad5e-61c427977ed5`: 各 `coat-lane-*.md` / `assets/manifest/coat-lanes/` とspeciesの新source/previewのみ。共通handoff/runtimeは親が更新する。
- macaroniには親が個別に開始画像revisionとFlow候補1本の所有を追加している。これは他laneの無制限Flow権限ではない。新しい許可・試行回数は最新台帳を読む。
- 現在のwhite guinea委任: `01a0e83b-9e19-7dd1-ad5e-61c427977ed5` は `parent-walk30-remaining26-v1/` の未消費D–Hを既存 `continue-D-H.json` に従って制作する。旧4pose準備のみの指示は終了済み。保留Bの修正は親の内部担当が別 `parent-walk-B-whisker-repair-v2/` を所有し、元remaining26の回数とは別に記録する。
- 当面の並行順序は黒rabbitのremaining26、white guineaの残りとB修正、degu blueのv6単pose試作、黒guineaのsinglepose v3準備。各段階の `.codex/qa/*-parent/authorization.json` が対象・所有・上限を定義し、この優先順位表自体は生成capを増やさない。成功した30poseと同色idleをまず一組ずつ親が統合する。失敗した4pose編集の同じ不具合を繰り返さず、単pose方式へ戻す。
- AnimalsDesktopとLINEは種・正式毛色の読み取り専用参照。prototype tintや別品種の形を、そのまま新しい正式色や同一baseと数えない。

## 初期計画の履歴 — 現行の実行指示ではない

2026-09-28。ユーザー指定: 色違いの制作をチャット01a0e2c7-ff04-7db3-bd76-7286a1a2d631へ依頼し、親チャットはAnimalsDesktopを参考にチンチラ等をFlowで制作する。今回は優先順位を確定する。

最新の進行指定: 指定チャットが旧待機指示で終了していたため、最初のGoalを**サンド1色の8コマ完成**へ限定して再依頼した。黒/チョコレートは次のGoal候補とし、現在のGoalへ混ぜない。新種はリアルな姿へ作り直すことをユーザーが追加了承。詳細な共通手順は `docs/workflows/animal-cursor-production.md` とプロジェクトSkillを使う。

2026-09-28 実制作更新: チンチラの写真風基準、Flow歩行2候補・待機1候補を作成。歩行v1は尾の端切れで棄却し、余白を広げたv2を採用候補にした。歩行22コマ917ms（GIF920ms）、待機96コマ4000msを64/96/256pxの透過APNG/GIFへ出力。12ファイルの再読込、hash、透過、時間、全118sourceコマの端を確認済み。詳細は `.codex/tasks/chinchilla-flow-20260928.md`。Windowsの歩行/待機切替は次段階。別チャットからサンド8コマ完成/Goal完了の報告も確認（親での再審査は未実施）。

### 初期計画の基準（履歴）

- 写真に近いリアル寄り。MofuMouseではデグーのFlow由来30コマと、ImageGen1回のブルー8コマ試作をユーザーが評価済み。
- 最新速度は元動画比2.25倍。30コマ版/APNGは556ms、ブルー8コマGIFは560ms。待機の呼吸やまばたきまでこの速度で加速しない。
- 現Windowsはアグーチ30コマ。色違い制作側は実行中exe・active.jsonを変更しない。
- 参照する現チェックアウトは `D:/開発/AnimalsDesktop`。`internal/catalog/catalog.go` には標準グレーのチンチラ、ゴールデンハムスター、砂色のマカロニマウス、チェスナット系うさぎのaccepted motionがある。一方、色候補のsourceVariantは採用済み動作と同義ではない。
- チンチラのsource-truthとaccepted frameを実見済み。色・耳・体形・ふさふさの尾・既存モーションの種類を参考にする。96×64の既存コマを拡大して新作の原本にしない。AnimalsDesktopへの書き込みは行わない。

### 初期5種の優先順（履歴）

| 優先 | 親チャット: 基本個体とFlow | 指定チャット: 色違い |
|---|---|---|
| 1 | チンチラ・スタンダードグレー。写真風の基準1枚→Flow歩行1本→良い1周期→透過APNG/GIF | 完成済みブルー8コマを保持。次にサンド8コマを同じポーズで1シート生成 |
| 2 | 同じチンチラの待機。足を固定した微かな呼吸、耳/鼻、少数のまばたき。歩行→待機の切替を確認 | デグーの黒・チョコレート。既存参照を確認できる色のみ、各1シートずつ。斑は後回し |
| 3 | ゴールデンハムスターの歩行→待機。短い足・短尾・低い体高が読めることを確認 | 基本個体の受け渡し後、チンチラのベージュ/エボニー。ホワイトモザイクは模様固定の別試験 |
| 4 | マカロニマウスの歩行→待機。太い短尾を維持。デグーからの種の置換で済ませない | ハムスター等の既存単色。長毛化や別品種化は色違いと分ける |
| 5 | ウサギの専用移動→待機。初期計画の立ち耳ルビーアイホワイトを維持し、他品種の耳・骨格を混ぜない | うさぎのオレンジ、次に模様付き。基準歩行の受け渡し後に着手 |

各種とも先に基本1色の歩行と待機を揃える。顔洗い・食べる・眠る等は5種の歩行/待機が揃った後。新種の優先はチンチラ→ハムスター→マカロニマウス→ウサギ。

### 最初の制作単位（履歴）

1. チンチラの写真風基準画像を1枚作る。既存チンチラ画像は種・毛色の参考であり、旧素材の画素感を移さない。横向き、全身、4足と尾が入る余白、現デグーと同じ光・自然な目。
2. 固定カメラ、その場で一定の歩行、単色の明るい灰背景、急な方向転換やズームなしでFlowを1本生成する。既存Flow手順を使い、モデル・長さ・出力数・表示コストを実行時に確認する。昔のモデル名や7クレジット表示を現仕様と決めつけない。
3. 保存したMP4の実fpsを確認し、中央の良い1周期を選ぶ。全種30枚や556msには固定しない。元fpsの連続コマと元の時刻を保存する。
4. 背景透過→共通位置と縮尺でPNG列→APNG/GIF。個別コマの脚パッチ・手描き修正・擬似上下運動を行わない。
5. 脚数、奥の後肢、接地、尾、毛先、輪郭の揺れ、最終→先頭、64/96pxを確認する。同じ問題が2候補で続いたら参照/指示を変え、無変更リトライを繰り返さない。
6. 速度は元速度を保存したうえでその種に合う値を設定する。デグーの2.25倍は現在のデグーへの採用値であり、新種へ自動適用しない。
7. 歩行が成立した同じ基準画像から待機を1本作り、歩行/待機の色・大きさ・基準位置を合わせる。試作で見せられる1種を完成させてから次種へ進む。

### 初期の書き込み担当（履歴）

- 親: `assets/source/animal-cursor-v2/<species>/`、`assets/source/video-cycle/animal-cursor-v2/`、そのpreview/manifest、Flowのブラウザー、Windows選択と再生処理。
- 指定チャット: `assets/source/animal-cursor-coat-batch-v1/`、対応preview、`assets/manifest/animal-cursor-coat-batch-v1.json`、`scripts/build_coat_batch_v1.py`、`.codex/tasks/animal-cursor-coat-worker-v2.md`。
- 採用済みデグー、ブルーv1/v2、既存原本、他担当の変更は上書きしない。参照は読み取り専用。別の色違い生成スクリプトを使い、親の速度変更やFlowスクリプトと編集を競合させない。

詳細な指定チャットへの依頼内容は `.codex/tasks/animal-cursor-coat-worker-brief-20260928.md`。
