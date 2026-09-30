# MofuMouse 写真風動物カーソルの制作Workflow

更新: 2026-09-28。これは実際に通過したデグー制作を基にした手順で、新種ごとの未検証部分も区別する。入口Skillは [mofumouse-animal-animation](../../.agents/skills/mofumouse-animal-animation/SKILL.md)。今後の順番と担当は [制作優先順位](../animal-cursor-production-priorities.md)。

## 段階0: 現状と担当を確定する

1. 最新のユーザー指定、AGENTS、対象manifest、直近の `.codex/tasks/animal-cursor-*.md` を読む。古い計画の「全コマを独立ImageGenで作る」「動画は参考のみ」は今回の採用方式ではない。
2. `assets/prototype/active.json` → 選択ディレクトリの `manifest.json` → `source_manifest` と辿る。実行中exeやPIDはその場で確認する。日付の新しさだけで候補を採用済み扱いしない。
3. 元画像・動画・選択コピーとhashを照合。存在すれば登録・再利用し、同じ画像を生成し直さない。外部の元生成画像が移動/消失していても、採用コピーが検証可能なら回復作業と新規生成を分ける。
4. 新種、毛色、動作、速度変更のどれかを選ぶ。速度だけなら段階6へ。現状確認だけなら生成しない。
5. 並行制作ではsource/preview/manifest/scriptの書込担当を明記。指定チャットの色違い担当はcoat-batch-v1、親は新種FlowとWindows。共通active.jsonを複数担当で変更しない。
6. 色違い担当の再開時と「次のreadyがない」判定前に、親 `animal-cursor-coat-handoffs.json` とworker側に保存したhandoffを照合する。`scripts/audit_coat_handoff_readiness.py` は元画像・8選択frameのhashと古い `waiting_for_base` を検出する。worker側の同期はworker自身が行い、完成済みsource/exportsを保持する。基準readyと色根拠確認済みは別条件として扱う。全種inventoryが未完成でも、基準と色を確認できた1組は制作を進める。
7. チャットへの送信成功やactive表示だけを着手完了としない。台帳の実更新、生成原本と選択コピー、prompt/provenance保存等の新しい成果を確認する。現在のユーザー再開指示がある場合、古い親返事待ちを停止理由として再利用しない。Goalのツール状態は実測を記録し、未達成をcompleteにしたり明示停止なしにpausedにしたりしない。
8. 全体の進捗照合には `scripts/audit_full30_coat_progress.py` を使う。基準動画30枚のready、旧8枚の色替え記録、専任laneの確認済み色/保留色、runtimeの実枚数とidle fallbackを分けて読む。色替え全体の分母が未確定なら完了率を作らず、元動画30枚が揃っただけで色替え30枚を完了扱いしない。監査の対応schema外の候補はunknownのまま残す。
9. 親が生成用hashを固定する前に、担当はPREPのQA追記も含めて完了・固定を報告する。固定後のQA追記は別ファイルへ保存し、source-lockやpromptを上書きしない。生成直前には実際の2参照とpromptを再照合する。2026-09-29のblack deguでは生成中にsource-lockのQA記述2点だけが更新された。元authorizationを保持し、2点の逆変換で旧byte/hashが完全一致するsnapshotとreconciliationを別保存して解消した。入力画像・promptの一致と記録のみの差を証明できない場合は、採用を保留する。画像を再生成して記録差を隠さない。
10. 出力のCALL receiptも担当の終端通知前には最終版としてpinしない。途中のsource画像レビューには原本hashと親側のreceipt観測コピーを使い、worker終端後の記録とは区別する。時刻は実測できる精度だけ記録し、保存処理の時刻をprovider返却時刻と呼ばない。blue v8では未確定UTC秒を日付へ訂正した差を旧byte/hashの完全復元で確認し、元の親HOLDを保存したままfinal decisionを別保存した。終端後の訂正はamendmentへ記録する。

### 実測済みの資産

| 項目 | 実際の状態 |
|---|---|
| 基準個体 | 写真風アグーチのデグー |
| 元歩行 | `assets/source/animal-cursor-v1/degu/flow/attempt-01.mp4`。1280×720、24fps |
| 採用区間 | 元動画1.5〜2.75秒、連続30枚、zero-based frame36〜65 |
| 透過source | `assets/source/video-cycle/animal-cursor-v1/degu/walk01.png`〜`walk30.png` |
| 最新manifest | `assets/manifest/animal-cursor-flow-walk-v3.json` |
| 現速度 | 元動画比2.25倍、54fps相当、30枚・556ms。画像は1倍/1.5倍版と同一 |
| ブルー色替え | ImageGen1回で8セル。`assets/source/animal-cursor-coat-sheet-pilot/degu/blue/v1/` |
| ブルー最新出力 | `assets/preview/animal-cursor-coat-sheet-pilot/degu/blue/v2/`、8枚・APNG556ms/GIF560ms |
| Windows待機 | 開眼2800ms/閉眼120msの既存2枚。Flow呼吸動画はまだ本体へ採用していない |
| 未検証 | 16/32セル一括、8セル×4シート間の色連続性、他種の新Flow歩行 |

ブルー8枚は元30枚の01/05/09/13/16/20/24/28を選んだ試作。コマ数が少ないことによる粗さはある。ユーザーがこの見た目とテンポを高く評価した事実と、全30枚の滑らかさを保持したかどうかを分ける。

## 段階1: 種の基準画像を作る

1. AnimalsDesktopの `internal/catalog/catalog.go` で対象種のsource/motionの所在と状態を見る。`acceptedMotionVariant` と単なる `sourceVariant` は区別する。現参照checkoutは `D:/開発/AnimalsDesktop`、読み取り専用。
2. 元絵を実際に開く。色名・耳・尾・体形・既存動作の種類を確認し、役割を記録する。96×64の既存runtimeコマを拡大して新しい写真風原本にしない。
3. LINE参照は採用色・模様の確認に使う。イラストの目の大きさ、太線、デフォルメは持ち込まない。根拠のない色や、違う種/品種を単なる色替えとして増やさない。
4. 内蔵ImageGen Skillに従い、1匹の写真風基準を1枚作る。横向き、固定視点、全身、尾・耳・ひげ・足の余白、自然な目、種に合う脚を指定する。
5. Flowへ渡すものは均一な明るい灰背景が実績のある出発点。透過identityが必要な場合はtrue alphaで別に保持する。透明入力をFlowに渡した際に黒茶のもやが出た履歴があるため、プレビューを確認する。
6. 原本を残し、repoへrevision付きでコピー。生成元/hash/prompt/参照役割/実寸/alphaを記録。64/96px、白/暗背景で種の特徴と欠けを確認する。

チンチラは写真風へ作り直す。密な灰色の毛、自然な丸い耳、コンパクトな胴、毛のある尾を基準とし、デグーの細い房尾を流用しない。新種の基準が成立するまで毛色を増やさない。

## 段階2: Flowで連続する動きを作る

1. 現在利用できるブラウザー操作Skill/ツールの仕様を読む。既存のFlowプロジェクト/タブを確認して再利用し、モデルや料金を記憶だけで決めない。
2. 開始画像を指定し、固定カメラ・右向き・同じ個体・一定縮尺・その場の繰り返し歩行を入力する。既定は1出力で小さく試す。過去は4秒/720pで試したが、現在のUIで有効な設定を確認する。
3. 表示モデル、解像度、長さ、出力数、表示クレジット、prompt、参照hashを記録して生成。追加購入やプラン変更をこの手順から推定しない。
4. 動画完成後、低解像度のストリーミング画面だけで品質を判断せず、元MP4を保存してfps/寸法/長さを調べる。元動画を保持し、選択コピーのhashを記録する。
5. 元動画を順に見て、4本の脚、奥の後脚の振り、接地、体形、尾、固定画角を確認。脚の複製、浮遊、テレポート、体の変形は不合格。
6. 同じ問題が2回続いたら入力画像、構図、指示のどれを変えるかを決める。無変更で何度も再試行せず、各候補の不採用理由を残す。

30実コマを目標にする場合は、24fpsなら1.25秒で支持足の位相が戻る指示を使う。4秒は3.2周期なので「4秒動画の先頭と末尾も同位相」を同時に要求しない。開始画像だけを指定し、1.25/2.50/3.75秒の反復から内部の連続1周期を選ぶ。実際の周期は生成後に検査し、秒数指定だけで30コマ達成とはしない。ウサギは両後脚で押す低いhopなど、その種の運動を維持する。

入力で奥後脚が腹の影に埋もれている場合は、プロンプト反復より先に開始画像の脚の分離を確認する。必要なら同じ個体の新しい単画像revisionをImageGen編集し、小さな視点差と前後の足配置で奥後脚を連続して見せる。原本保持・合計2後脚・手描き修正禁止は変わらない。

待機は歩行が決まった同じ個体で別生成する。足を固定した小さな呼吸・鼻/耳・まばたきが中心。開始/終了を同一画像にしてもループは保証されない。デグーidle案02は灰背景が改善したが、瞬き回数と体の沈み・終端差があり、完成ループとしては未採用。

プロンプト例はSkillの [prompt patterns](../../.agents/skills/mofumouse-animal-animation/references/prompts.md)。既存の成功promptは動画に隣接する `.prompt.txt` に残す。

## 段階3: 元動画から良い1周期を選び透過する

1. 動画のfpsと時刻を取得し、頭/胴の位置・各足の位相・終端差を比べる。見た目が近いだけで足の接地位相が違う区間を繋がない。
2. 1周期の開始・終了、元のzero-based frame番号、原動画hashを記録。元のfpsで連続コマを切り出す。新種へデグーの1.5〜2.75秒や30枚を固定適用しない。
3. デグーではローカルIS-Net general-useでalphaを推定し、元RGBをそのまま保持した。元デコードとsource RGBAのRGBが全枚一致することを検証する。
4. 毛先、尾、ひげ、脚の間を原寸と64/96px、白/暗背景で確認。モデル出力をalphaにしただけで高品質透過が保証されたとは扱わない。
5. 解像度変更やtrimは全コマで共通範囲を使う。デグーは1280×720を352×198へ縮小し384×256へ(16,29)で配置。これを64/96px高等へ縮小した。
6. コマごとのbboxへ毎回ぴったり拡大しない。自然な上下動と、生成による体形/位置の崩れを区別する。脚の描画や局所合成で修復しない。

デグー実装: `scripts/build_flow_walk.py`。`--pilot`は01/11/21の透過比較のみ。元動画範囲、30枚、model、出力v1が固定のため、新種へ無変更実行しない。

新種の周期候補を調べる補助は `scripts/inspect_flow_period.py`。まず `build_flow_animal.py --inspect` で元fpsの全コマをdecodeし、configとvideo/cacheのhashを一致させる。`--roi` は320×180診断画像上の脚領域、`--target 30` は選択枚数。出力の `last_selected` と比較用 `withheld_next` を混同しない。lag最小値は採用判定ではなく、全4肢の支持位相と元コマを目視する候補情報。フェレットattempt02では小さな終端差でも実際は27コマ周期で、30枚の末尾3枚が次周期へ入り不採用となった。出力JSONは新しい名前で保存し、以前の根拠を上書きしない。

ハリネズミattempt06では240枚すべて異なるhashでも、長い区間で足の支持位相が止まっていた。最小差分の30窓は静止を選んでしまい、動いている窓も全4肢が一巡する前に30枚を超えた。異なるPNGが30枚あること・開始終了画像が同じこと・動画が長いことは、30実ポーズの歩行周期の証拠ではない。最初に全動画の各足の動く区間を確認し、静止窓は除外する。configを検査後に更新する場合は、診断JSONがhash参照する選択前configをbyteコピーで保持する。

透過後の寸法・色・終端差は `scripts/inspect_motion_geometry.py --manifest ... --output ...` で共通座標の96px高PNGから測定できる。alpha128の胴maskを7px収縮/膨張し、面積平方根比・中心・平均RGB・暗背景合成MAEを記録する。コマ別位置合わせは行わず、source hashを確認して新規JSONだけを出力する。既存ferret診断値との一致を確認済み。数値閾値だけで脚やループを合格にしない。

## 段階4: アニメ出力と視覚確認

1. 正規化PNG列を保存。フレーム順、source hash、サイズ、alphaを検査する。
2. APNGはRGBAのまま、blend=source相当の全フレーム置換、無限loop。GIFは全コマ共通パレット255色＋透明index、alpha閾値128、disposal=2。半透明の毛先はAPNG/Windowsを基準にする。
3. コマの表示時間は累積時刻の差で作る。毎コマの時間を単独で丸めると周期がずれるため、累積境界を丸める。
4. GIFは10ms単位。20msより短いdelayを避ける互換出力では、必要な場合だけGIFを間引き、枚数と周期を別に記録する。現在のアグーチはAPNG/native30枚556ms、GIFのみ28枚560ms。ブルーは8枚のまま560ms。
5. 全てのGIF/APNGを再読込して、実コマ数、各delay、合計時間、サイズ、透明部分を検査。ファイルを保存しただけで成功にしない。
6. 原寸の連続図、64/96px白暗、アニメ再生、ループ終端を確認。両後肢が動くか、余計な足がないか、毛色がちらつかないか、尾や耳が切れていないかを見る。IoUは位置ずれの補助値。

ここで64/96pxはアプリのcanvas高さ。横幅64pxに灰背景の全動画を縮めた画像はruntimeサイズの証拠にならない。共通crop→384×256配置→96×64または144×96の実際の出力条件で比較する。4セル色替えは同じ歩行内だけでなく、既存の同色idleとの胴体色と明暗ボードも比較し、歩行/待機切替で色が変わる候補を止める。
7. 生成済み・候補・ユーザー評価済み・runtime採用済みを別に記録する。写真風の画質不満が出たら、元動画→alpha→縮小→表示を辿る。

Windows診断中に画面構成が変わるとoverlayのHWNDも変わる。`<report>.live.json` を一度取ったまま使わず、再作成後の現在windowを検証する。`--smoke-display-rebuild` はこの経路を明示的に通す。旧windowの遅延load失敗を現行windowの失敗として扱わない一方、現在ownerのload/rendererエラーはfailを維持する。検査は旧HWND無効・新HWND有効・click-through/no-activate・下のwindowへのhitを分けて記録する。有限時間の描画coverage不足は、実際の次周期を上限付きで追加観測し、seenを捏造しない。

## 段階5: 元の参照ポーズから色違いを作る

### 既存Flow動画を色替えする試験経路（2026-09-29、受入前）

- Omni 1.1 Flashの動画素材編集を使う。新規歩行の再生成と同一視せず、採用済み動画を動作・位置・縮尺の参照、個別ImageGenの単姿勢を毛色だけの参照に分ける。元動画・元画像・prompt・生成回数を別々にhash記録する。動きの保持は指示であり保証ではない。
- 元動画がプロジェクトに残っていれば、保存済みFlow asset URLで正しい版を開き、元promptと照合して使う。同名の失敗動画と混同しない。必要なら採用元の版番号をasset名に付ける。再アップロード時の権利同意を自動承認せず、既存素材を直接選択できるか確認する。
- RGBA毛色参照の黒い見え方はalphaを確認する。Flow用には元画像を保持したまま全canvasを中立灰色へ通常alpha合成した転送用PNGを作れる。動物の描き替え、色調補正、bbox fittingはしない。転送PNGは別hashで記録する。
- 現UIでは素材選択に動画と画像の両方を入れると「素材に基づく長さ・4秒」と表示された。720p、16:9、x1、Omni 1.1 Flashを確認し、表示費用をその場で記録する。今回の動画編集は20creditsで、開始画像だけの7creditsとは異なる。ONE call/retry0の段階を保存してからGenerateを1回だけ押す。
- downloadイベントの観測timeoutはダウンロード失敗ではない。実際の完成ファイルとFlow結果URLを確認し、重ねて生成しない。元720p MP4を保存し、実fps/寸法/全frame数をdecodeして検査する。
- チンチラ最初の試験は `assets/source/animal-cursor-v2/chinchilla/beige/flow/walk-color-edit-01.json`。1280x720、24fps、96frame。灰色元の26..55という30frame範囲は比較仮説であり、色替え結果の受入ではない。全肢の位相・奥後脚・終端26対56・毛色・共通座標の寸法を再検査する。個別コマの位置合わせやRGB後処理で不一致を隠さない。
- この最初の試験では実際に脚位相が変わり、灰色と同番号frameの対応が失われた。ベージュ26は手前後足が後方接地、56は次の前方振り中で、選んだ30frameは閉じた1周期にならなかった。冒頭の静止に近い区間も含む。30枚の異なるPNG・alpha・1250ms・export読戻しの合格だけで採用しない。元動画編集を全色の確定量産経路と呼ばず、同じ設定の自動再試行は行わない。


現行は元frameを1画像1匹1ポーズで直接編集する方式を既定とする。2026-09-29の最新ユーザー添付AGENTSにより、既存4poseの2×2編集は段階5Aの限定試験として許可されている。以前の全面禁止は更新済みであり、4pose試験の方針確認は不要。ただし各試験の対象・出力先・新しいcall上限は親が明示し、消費済みcapや旧8セル生成を再開する根拠にはしない。旧ブルー/サンド8セルと過去の4コマ試験は履歴として保持する。

新規walkは採用済み連続30実frameの出典を固定し、まず01/11/21等の離れた個別poseを限定call数で試す。pilot合格後、親が残りposeと新しいcall上限を指定する。全30枚の元参照がない場合は単一poseの試験までとし、8枚をwalk完成扱いしない。全poseに共通のcrop/scale/anchorを用い、採用idleとの色・寸法、全30枚の連続性と最後→最初を検査する。

以下の8選択pose・配置・delayは旧デグー試作を再現するための履歴例であり、新規30frame制作の選択枚数や生成指示ではない。既存PNGを組んだ技術比較板は利用できる。複数pose生成の入力に使えるのは、現ユーザー指示で許可された段階5Aの4pose試験だけで、8pose入力は新たに許可されていない。

1. 採用済み動作から8コマを選ぶ。デグーの実例は01/05/09/13/16/20/24/28。別動作では足の重要位相を見て選び直す。
2. 全コマを同じcrop/scale/anchorで2列×4行1536×1024へ配置する。1セル768×256。文字やセル枠は生成入力に焼き込まない。
3. デグー実例では共通crop=[0,105,1280,651]、内容525×224、セルoffset=(121,16)。この値は既存コマを測って得た結果で、新種へ固定しない。
4. 参照シートを実際に表示。コマごとの元PNG/hash、番号、セル境界、座標変換を記録する。
5. 比較シートから各セルを個別参照に切り出す。内蔵ImageGenは1回につき1匹1ポーズで編集し、参照1は正確なポーズ、参照2がある場合は毛色だけと明示。同じ個体・脚・尾・配置を保持し、検証済みの毛色だけ変更する。
6. 各生成原本を保持しrepoへコピー。8ポーズには8枚それぞれの生成出典/hashが必要。融合、足/尾の欠落、偽の市松模様があればそのポーズだけ作り直す。採用済み旧シートには実際の1回生成出典を保持する。
7. 全ポーズに同じcrop/scale/anchorを適用し、コマごとのズレを個別補正して隠さない。透明度に問題があれば原RGBを保持した既存alpha抽出経路を使い、方式とモデルhashを記録する。手足の描き足しはしない。
8. 元ポーズとの比較と段階4のQAを実施。ブルー実例は配置がよく保たれた一方、毛並み・顔・後脚周辺の陰影も少し描き直されている。完全な色だけの変更とは報告しない。
   透明PNGの周囲に発光や色が見えた場合は、原本を白・暗背景へ正しくRGBA合成してから判断する。alpha=0内のRGBや低alpha毛先だけでは可視ハローの証拠にならない。2026-09-29のferret champagne ID47では元viewerの見え方を不良と誤認し、実64/96/256px合成と独立QAで訂正した。初判定・訂正根拠・原本hashを残し、表示上の誤認を消すために画像を加工しない。
9. 元コマ間の時間差を8枚へ割り当てる。全コマを同じ時間にしない。現在の2.25倍速では[74,74,74,56,74,74,74,56]ms。
10. 同種の単色から進め、白斑等は模様の位置も確認する。新種/品種/長毛化を単なる色替えとして扱わない。

#### 単画像の色替えで確認した制約（2026-09-29）

- 入力1は正確な元pose、入力2は色の参照と役割を限定する。guinea cream idleでは無地の小さな色swatchによる2参照で8poseの形・目・四足を保った候補が得られた。これは全種・全色への成功保証ではなく、個別pilotを通す。動物の姿勢を含む別画像を色参照へ追加しない。
- 同じ色のwalk内だけでなく、採用済みidleとの切替を高さ64/96/256・共通配置で確認する。ferret champagneの旧walk3枚は互いに安定していても、idleより背・尾・足が明るすぎた。腹/脇の明るい毛だけの参照とlight pawsの指定が全身を白くする場合、採用idleから部位別の色を計測した無地の色帯を参照し、背・脇・尾・足/顔の役割を文章で指定する。新しい1枚で検証するまで残りの量産を開始しない。色帯は参照資料であり、動物への塗り重ねや生成後の局所着色ではない。
- 部位別の比較は同じ絶対座標の矩形ではなく、各motionで同じ解剖学的部位を選び、元色のwalk/idle差も比較する。ferret champagne v3の顔は元sableでも約-9luma、編集後約-10lumaで、主に顔向きと影の差だった。一方flankは元色+2.7に対して編集後-11.9で、色編集に由来する差として保留した。数値の完全一致を求めて自然な陰影まで消さない。
- ループ境界も同じ基準で見る。rabbit black idleの08→01では候補RGBが[-15,-10,-9]、同じ部位の元whiteが[+1,+1,+1]だったため、陰影では説明できない色の点滅としてfull8を保留した。個別poseの合格だけではsequence受入にならない。旧01–03と新04–08の手法差は保持したまま、必要な1枚の新revisionから検証する。
- 古いpromptと新しいpromptが混在するsequenceでは、毛色以外の変更指示も比較する。rabbit black旧01/02は茶色い虹彩、新06–08は元whiteのruby眼で、開閉順が合っていてもH96で色の違いが見えた。現在のsource-ruby維持方針を明示し、修正コマの前後両側を確認する。01だけの修正で02の不一致まで解消したとは扱わない。
- 元poseの細部を実画像で読んでからpromptを流用する。rabbit白idle001には実alphaを持つ長い放射状ヒゲがあり、後半pose用の「元から短い下向きのヒゲだけ」という記述は不適切だった。元の長さを保持するか、自然な短いヒゲへの修正を唯一の意図した例外として認めるかを生成前に決め、旧promptと理由を残す。見えるヒゲを不可視RGBと決めつけず、RGBA合成で確かめる。
- 無地paletteの数値変更が生成毛色へ線形に反映されるとは限らない。degu blueの背の色帯を約+15明るくしても生成背は約+3lumaにとどまり、尾は逆に暗くなった。同じ値調整を繰り返す前に、色を示す文章・光の指定・実参照の役割を確認する。次の試験はpaletteを固定して色説明文だけを変える等、変数を限定し、原因は検証前の仮説と記録する。
- generation前にprompt本文と実際の参照配列を両方確認する。「元pose＋無地swatch」の契約へ、QA用の採用idle全身絵をImage3として紛れ込ませない。準備資料を訂正した場合は旧hash、訂正理由、生成前でcall0だったことを記録する。
- 透過済みposeを色替えする場合、prompt本文も`transparent_background: true`も透明RGBAを指定する。`plain neutral gray matte`等の背景追加指示を同じpromptへ混ぜない。チンチラbeigeの準備でこの矛盾を生成前に発見した。背景を付ける指示だけを新revisionで直し、元prompt/hash・call0を保持する。見た目だけで不透明と決めず、戻った原本のalphaと白/暗背景合成を確認する。
- 全体の位置ずれは、相対的な体形崩れと分けて診断する。チンチラbeige01/11の未補正ずれは約22/14pxだったが、両方へ同じy=-18px（768×512上）だけを適用した技術比較では、H96の足元残差は各±0.75pxだった。これは次の離れたposeを試す根拠であり、全30枚の安定やproduction補正の受入ではない。各コマへ別の移動量や倍率を当てず、同じ変換を残りのposeと連続再生でも検査する。
- 元poseのヒゲも保護対象として実表示で比較する。guinea creamの追加02/03では、body寸法・毛色・脚・外縁alphaが良好でも、元goldenより長く明るい複数のヒゲがH64/96で見えた。複数候補に同じfanがあることを「連続性pass」の根拠にしない。元画像と候補を同じH64/96/256・白/暗背景へRGBA合成し、採用pilot/idleとも比較する。低alphaの不可視残留と、実際に目立つ追加strandを分ける。materialな反復を見つけたら次callを止め、inflight原本と実call数を保存する。生成後にヒゲを塗り消さない。rabbit05-r3では正確な元白pose＋同じflat色参照に戻し、元の微かな短い暗いヒゲだけ保持する1回の新revisionが合格したが、他種への成功保証にはしない。
- 生成原本が1537×1023等、想定1536×1024と1px違う場合、元画像を保持する。右端1列のalphaが全て0と検証できる場合だけ、原点固定で透明列を除去し下へ透明行を追加した1536×1024へ整え、その後全画面を同じ倍率で縮小する。guinea cream idleのP3/P5/P7でこの補正を記録し、独立再構成がpixel一致した。切り捨てる列に動物が触れていれば保留する。各画像をbboxに合わせて拡大/再配置したり、縦横別倍率で合わせない。
- provenanceには実際にtoolへ渡した入力path/hashと、返った原本path/hashを別欄に保存して実ファイルと再照合する。変数の上書きで出力pathが入力欄へ入る誤りを検出した場合、実呼出の証拠から訂正し、失敗画像・累計call数・判定を保持する。

現在のLuna向け実行手順は `.codex/tasks/animal-cursor-coat-ten-species-brief.md` と担当台帳。旧 [coat-sheet-workflow.md](../coat-sheet-workflow.md) の一括生成案は履歴参照のみで、新規実行指示ではない。

#### 透明入力と白合成入力の診断（2026-09-29）

- ヒゲの輪状化が複数の色・poseで繰り返す場合は、元decoded RGB、IS-Net alpha、実送信入力のRGB/alpha、正しい白暗合成を比較する。degu01/02では低alphaの房状縁がalpha無視の表示でU字に見えたが、成功した01にも同様の縁がある。providerがalphaを無視したという観測はなく、原因断定や同じ条件での無制限再試行をしない。診断は `.codex/qa/degu-pose02-whisker-input-diagnosis/`。
- 黒idle01を正しく白合成した参照で1回編集した試作は、真の透明背景と非輪状ヒゲを得た一方、元の共通canvas比で高さ+13.6%、面積+31.1%、重心が約20px上へ移動したためHOLD。透明・顔の改善だけで採用せず、必ず元寸法/配置も確認する。これは対照実験のない1例で、白合成の成功法則ではない。個別bbox fitや位置補正で幾何の失敗を隠さない。親決定は `.codex/qa/degu-black-idle-white-input-v2-parent/source-decision.json`。
- 続く黒idle v3では、元の透明canonical参照へ戻し、細く疎な元ヒゲと透明余白・寸法・配置の保持を明記した別2callで開眼01/閉眼51をsource採用した。01の面積差は+1.96%、IoU0.96812、51はIoU約0.9643で、白合成試作の拡大は見られなかった。ヒゲ密度と歩行から待機への腹の明暗差はwatchとして残す。親決定は `.codex/qa/degu-black-idle-transparent-v3-parent/source-decision.json`。参照背景とpromptを同時に変えた限定結果なので原因や他poseへの成功を断定せず、残り6poseと全8poseの色・寸法・継ぎ目を別に確認する。
- Blue walk白表示入力v11/v12は各1callでpose02/01を作成し、共通canvasのIoUはともに約0.981、寸法・脚・ヒゲ・alphaは通過した。しかし同一色カードと文面でも隣接poseの背色中央値が99/95/95と115/106/104に分かれ、親と独立QAで色連続性HOLDとした。旧accepted01は保持する。白入力は万能解ではなく、形状改善と色の安定を別に判定する。同じ色値/promptの反復は止め、別の明確な仮説を作ってから次段階へ進む。親決定は `.codex/qa/degu-blue-white-view-paired-v11-v12-parent/source-decision.json`。
- CALLの回数欄は意味を確かめる。black phase02/10/16の `generation_calls` は段階累計1/2/3であり合計6ではない。原本・pose・PRECALLとの対応から各1call、段階3callと数える。中止前に書いたPRECALLだけでは実送信に数えず、返却原本と実送信証拠を区別する。
- 凍結promptをfunctionsから送る場合、PowerShellで `Get-Content -Raw -Encoding UTF8 '<prompt-path>' | ConvertTo-Json -Compress` を実行し、成功したexec結果の `output` を `JSON.parse` して文字列へ戻す。実終端LFも保持できる。functionsのV8では `atob` やNodeの `Buffer` を前提にせず、独自Base64 decoderを足さない。送信前のparse失敗はprovider未呼出しと確認できる場合だけcall0。送信後の曖昧な応答は未消費と推測せず、同じ認可で再送しない。
- 2026-09-29の黒walk10/24は白RGB表示の正確なpose＋黒単色カードで各1callの単体source検査を通過した。共通canvasのIoUは約0.977/0.971。寸法・4足2後脚・細い分離ヒゲ・真alphaを確認した限定結果で、未制作25poseや全30の連続性を保証しない。pose10固有の長い胸方向ヒゲを保護する文章は他poseへそのまま移さない。親決定は `.codex/qa/degu-black-single10-white-view-v4-parent/` と `degu-black-single24-white-view-v5-parent/`。
- Bluev13でaccepted01の全身画像を色参照にすると色は近づいたが、正確なpose02のIoUが約0.875、重心が約9.7pxずれ、ヒゲも再び崩れたためHOLD。全身参照は色だけの権威として指定しても形を混ぜる可能性がある。次候補の「白表示の正確なpose＋輪郭を含まないnative毛色部分カード」は未試験として区別し、同条件を反復しない。Whitev3はヒゲ・寸法・白毛alphaを通した一方、選択外見に反して尾tuftだけ暗いまま残った。単一部位の修正では対象外の全身色・形・alphaも比較し、全体再生成や手RGB修正へ広げない。
- 黒idleの選択8pose/4000msはsource採用済み。APNG再decodeの画素/時間一致と、親IABでの白暗64/96・瞬き・96→01の再生を別々に記録した。PNG列の画面statsはnativeAPNGのframe検出ではない。27/87→93の軽微な明度差はwatch。96の上方12pxはalpha1で実表示に断片が見えず、alpha1 bboxの伸びだけをサイズ破綻と判定したり原本を掃除しない。親決定 `.codex/qa/degu-black-idle-last4-v5-parent/source-decision.json`。素材採用はElectronでの歩行↔待機確認やruntime採用とは別。

### 段階5A: 1画像4コマの色替え試験（明示された限定例外）

2026-09-29の最新ユーザー添付AGENTSは従前のAGENTSを置き換え、個別poseを既定としつつ既存poseの2×2編集試験を明示的に許可した。これは現在チャットの指示に基づく更新であり、過去の手順だけから許可を推測したものではない。各試験は親が参照/hash・対象4pose・出力先・call上限を記録して開始する。消費済みの旧capは復活せず、8以上のsheetや全色量産は許可されない。将来のチャット指示が再びmulti-poseを禁じたら、その指示を優先する。

初回履歴はguinea black idle。再開する試験は同じ種・同じmotionの4つの実poseを対象に新しい段階として指定する。単一frame方式との比較が目的であり、採用前に全色へ拡大しない。現在の個別生成が進行中なら、その担当・出力を上書きせず別所有領域で準備する。

1. 採用済み同一motionから4つの異なる実frameを選び、元PNG/hash/動画frame index/timestampを記録する。idleは開眼・閉眼を含める。4つのポーズを新たに想像して描かせない。
2. 既存PNGを同じ縮尺・キャンバス・足位置のまま2×2へ配置し、四象限の境界座標と逆変換をJSONへ保存する。共通余白で耳・尾・足をセル内に収め、ラベル・罫線は焼き込まない。この参照配置は機械的な合成であり、動物の描画・脚の修正ではない。
3. 入力をview_imageで見てから内蔵ImageGenにeditとして渡す。参照1は編集対象の2×2、参照2は毛色だけ。順序、四象限、姿勢、体形、脚の数、尾、向き、目の開閉、足位置、縮尺、余白を保持し、その段階で確認済みの対象毛色だけ変更する。透明背景を要求し、セル枠・文字・影の追加は禁止。
4. 原本を保持して専用laneの新revisionへコピー。生成1回・出力1枚・セル4個と明記する。入力と出力のhash、prompt、実際の画像寸法、実行時刻を保存する。
5. 生成画像の実寸から固定の2×2境界で切り出す。全セルへ同じ変換のみ適用する。セル別の位置/縮尺補正で失敗を隠さない。セル越境、欠落、重複pose、偽市松、余分な脚、切れた尾、顔の変形があれば不合格。
   透過の判定は必ず実RGBAを白/暗背景へ正しく合成して行う。white guinea walkの初回4pose原本はviewer上で広い光背のように見えたが、正しい合成では現れなかった。不可視RGBだけを背景不良と誤判定して再生成・alpha切削しない。一方、この試験では細い元ヒゲが可視の明るいfan/太い束へ変わったため、脚位相・寸法・色方向が良好でも4候補はHOLD。独立した微細部修正の新revisionが必要な場合は、原因・変更範囲・元hash・新capを明示し、元絵を保存する。
6. 原寸・64/96px、明暗背景、4コマ連続再生で確認する。元frameと輪郭/接地/胴体寸法、同色間の白バランス/毛色を比較し、単一frame試験より品質が落ちていないか記録する。比較用APNG/GIFは4コマ試験と表示し、全idle/walk完成に数えない。
7. 不具合が出たら原本と不採用理由を残して停止する。明確な変更仮説と親の新しいbounded指示がある場合だけ次の限定試験へ進む。この手順文からretry権限を推測しない。同じ破綻が続く場合は個別編集へ戻す。合格後に親が方法採用を記録するまで他laneへ横展開しない。
8. 30実ポーズのwalkに採用できた場合は、4×7組＋残り2コマを個別編集する案から始める（初回生成9回、修正は別計上）。組境界と最後→最初を通してQAする。4コマ×8組から2枚を重複扱いで混ぜることや、4枚を複製して30扱いすることは禁止。実generation countと採用pose countは別々に記録する。

2026-09-29の過去の検証記録には、guinea黒idleの2枚/2実生成による8poseと方式評価pass、およびguinea walk30のid1/10/16/24試験への拡張がある。ただし現在の黒idle採用は0であり、過去の方式評価と現在の素材採用を混同しない。細かな毛・ひげ・鼻の再描画も残るため完全な色のみ変更とは表現しない。既存原本・検査を保持し、新規callは現在の親段階に結び付ける。

### 受入済み1motionだけをローカルElectronへ反映する

`electron-prototype/scripts/replace-reviewed-motion.py` は、新しい24fps・30連続実frame・1250msのreviewed_source_ready歩行を、指定variantのwalkだけへ反映する。既定はdry-run、`--apply`で変更。既存18種類や同variant idleを再importせず、旧runtime PNG/manifest/snapshotをqa/motion-replacementsへ保存する。変換は旧表示のbody面積/中心/足位置に合わせたmotion共通scale/offsetだけで、RGBや脚を描き変えない。デグー556ms等のリタイムにはこの限定helperを無変更で使わない。

梱包は `node scripts/package.mjs win32 x64 <local-label>` で既存版と別directoryへ出せる。source gate→限定runtime差分→チェック→actual packaged smoke/catalog/native→通常起動を順に行い、公開版は別扱いにする。PowerShell helperの成否はJSON passまたは例外を使い、成功時未設定のLASTEXITCODEを失敗と誤判定しない。終了済みsmokeのnative window handleは再利用せず、現在liveな診断processから読取確認する。

Catalog撮影中にもdisplay rebuildが発生する。2026-09-29に旧windowの`capturePage()`が`UnknownVizError`となる例と、unloaded時のearly returnで19種類を検査しても40/41枚になる例を確認した。`capture-current.cjs`は現在window・selectionのdecoded/paint済みsnapshotを待ち、capture前後のowner確認を行う。retired windowの結果だけを破棄し、現在windowのcapture失敗はそのままfailにする。待機にはdeadlineとretirement上限を設け、async hookも上限外で待たない。`--catalog-smoke --smoke-display-rebuild`は最初の撮影windowを意図的に置換し、同じcaseが別windowから保存された証拠と全41captureを要求する。失敗reportは保持し、単に撮影枚数gateを下げたりUnknownVizErrorを全て無視したりしない。

### 受入済みの色違いを1種類だけ追加・更新する

`electron-prototype/scripts/integrate-reviewed-coat.py` は、walkとidleを揃えた1色を対象とする取り込み処理。契約は `.codex/tasks/coat-targeted-integration.md`。2026-09-29にローカル検証を完了した。11件の単体テスト、既存catalogと19件の実参照の読取検証、中断復旧時の改変とWindows junctionによる保存先逸脱の独立検査が合格。最終証拠は `.codex/qa/coat-targeted-integration-review/parent-verification.json`。実素材へのstage/apply、更新後のWindows梱包・Mac共通ソース更新はまだ行っていない。

1. Receiptに現catalog/snapshotのSHA、対象IDと明示的なadd/replace、walk/idleそれぞれの基準参照・原本・出典・実時間・H64/H96 APNG・motion共通の表示変換を記録する。歩行は30連続実pose、デグー556msと他種1250msを区別する。待機は基準の実pose順・元ID・実時間を保ち、8枚だから均等間隔とは扱わない。
2. `payload-hash --receipt <repo-relative-json>` でreviewを除くReceiptのcanonical hashを出す。親のsource/export/presentation目視検査JSONにこのhashと対象ID、証拠のpath/hashを記録し、その検査JSONをReceiptへ結び付ける。検査済みという文字列だけでは受け入れない。
3. `stage --receipt <repo-relative-json> --output electron-prototype/qa/coat-stages/<new-stage>` で隔離したcatalogを作る。既存stageの上書きは不可。現在のcatalogや保存設定を変更せず、元PNGを上書きしない追加番号で配置する。無関係な種類のレコード・順序・既定IDを維持する。
4. 検証を通したstageだけを `apply --stage electron-prototype/qa/coat-stages/<stage>` で反映する。元catalog/snapshotのbackupとjournalを保存し、画像コピー後にcatalogをatomic置換する。入力・検査・stage・現在catalogのどれかが変わっていれば止める。
5. 中断時は同じstageの `recover --stage ...` を使う。journalと既知の旧/新hashを照合し、catalogに対応するsnapshotへ戻すか完了させる。不明な状態を推測して上書きしたり、残った旧PNGを削除したりしない。
6. Apply成功はsource取り込みまで。別ラベルのWindows packageを作り、混在隊列・個別選択・walk/idle/色切替・寸法・速度・クリック透過のnative/visual確認、設定保持、Mac共通ソース更新を行ってから通常版へ切り替える。公開Actionsの回数制限とは別のローカル工程であり、追加公開の許可にはならない。

各コマのbbox fitting、RGBの塗り直し、脚の補修、複製や補間による30pose化は行わない。CLIは実在確認したPythonから実行し、引数のpathはrepo相対にする。テスト用rootや合成fixture画像は本番素材の受入証拠に使わない。

### 既存色の待機だけを更新する

`integrate-reviewed-coat.py` の明示的な `operation: "replace-idle"` は、すでにcatalogへ存在する同じ色のidleだけを差し替える。2026-09-29に独立レビューで検出した元動画時刻の重複を拒否する修正後、親側の23テスト・実入力20検査・独立再現検査が合格した。最終実装証拠は `.codex/qa/coat-idle-integration-parent/verification.json`、台帳は `.codex/tasks/coat-idle-integration-implementation.md`。既存の `add` / `replace` は引き続きwalk30とidleの両方を必要とし、walkの省略からidle-onlyを推定しない。

- Receiptのmotionsはidleだけ。基準source・全pose・実duration・原本/provenance・透明APNG H64/H96・motion全体に共通の表示変換・親のhash-bound検査を通常と同じように確認する。
- 既存walkの全PNG、出典、時間、表示変換、他の動物、並び順、既定選択を保つ。idleFallbackがwalk PNGを参照している場合もそのPNGを除去しない。新しいidle PNGは追加番号で保存する。
- 隔離stage、apply前の元hash照合、差分のReceiptからの再構成、中断復旧を通す。実素材の書き出し・stage・apply・梱包・native確認はそれぞれ別の証拠として残す。
- デグーサンドの暫定walk8を残してidle1静止画をidle8/4000msへ改善する用途でも、full30の色違い完成とは数えない。新しい色をidleだけでcatalogへ追加する用途には使わない。

2026-09-29にデグーサンドへ初めて実適用した。source/export/presentation検査をhash付きReceiptへ結び、隔離stage→apply後、既存2764 PNG・walk・他18種類・既定/順序の不変とidle16 PNG追加を確認した。ローカルWindows `sand-idle-v3-20260929` は26テスト、砂色全歩行/待機コマの描画、全19種類41capture、display再構成、混在10匹、nativeクリック透過と設定保持を通過し、通常起動へ切り替えた。Mac共通ソースは2780 PNG版へ更新、Mac実機と追加公開は未実施。証拠は `.codex/qa/degu-sand-idle-runtime-parent/RESULT.json`。元のworker draftに残る取り込み前catalog hashは書き換えず、親の最終Receiptを別ファイルに保存する。

## 段階6: 再生速度だけを変える

1. 現在のmanifestの速度とユーザーの相対指定を確認。「さらに1.5倍」は現在速度×1.5。デグーでは1→1.5→2.25になった。
2. 元動画の時刻を基準に新しい累積境界を作り、差を各delayにする。`round(frameIndex*1000/(sourceFPS*speed))`を基準に、ループ末尾を含める。
3. 既存PNG/decoded RGBAが変わっていないことをhashまたはbyte比較する。生成・描き直し・補間の新規追加は行わない。
4. 出力先とmanifestを新revisionへ。過去のGIF/APNG/exeを保持。歩行の変更を待機へ波及させない。
5. API/nativeの最短delay制限も確認する。現在は10ms更新タイマー、入力delay許容10〜2000ms、実際の54fpsは18/19ms。短い時間を無理に0msに丸めない。
6. フレーム境界の直前/直後、複数ループ、停止再開を確認。GIF/APNGを再読込し、形式による丸め差を記録する。

実行済みコマンド（利用できる実Pythonを選ぶ。Store aliasを使わない）:

```powershell
# プロジェクトrootから。これはデグー専用の既存revisionを再出力する。
& $taskPython scripts/retime_flow_walk.py --speed 2.25
& $taskPython scripts/retime_coat_sheet_pilot.py
```

`--speed 1.5`ならFlow v2、`2.25`ならv3。汎用の任意速度指定ではない。新しい速度値や新種を扱うときはscriptの実装を先に確認する。

## 段階7: Windows試作へ組み込み、実行を確認する

1. 元source、output PNG、manifestのhash/枚数/時間を確認する。異なる画像への上書きを拒む既存prepare helperの保護を外して進めない。
2. 使用manifestを指定してruntime PNGとactive.jsonを準備する。
3. Goテストを実行し、稼働中exeと違う名前にビルドする。既存プロセスを停止する前に旧exeと退避コピーのhash一致を確認する。
4. 試作exeの正確な絶対パスに一致するプロセスだけを終了する。他の動物アプリや同名の別配置は停止しない。
5. 新exeを64/96px、各7秒のsmokeで実行する。ExitCode=0、walk_frame_count=30、期待周期、全walk＋open/blink描画、exit_hotkey_registered=trueを読む。JSONキー名は実装と確認する。
6. 合格後にcanonical exeへコピーし `scripts/run_pet_prototype.ps1` で起動。PID/Path/Responding/hashを再確認。自動起動やユーザー設定は変えない。
7. クリック透過、左右反転、停止再開、画面端と実サイズを必要に応じて確認する。APIの描画成功と画面の見た目は別の証拠。古いデスクトップ画像が返る撮影は「現在のスクリーンショット」として使わない。

```powershell
& $taskPython scripts/prepare_pet_prototype.py --walk-manifest assets/manifest/animal-cursor-flow-walk-v3.json
& $taskGo test ./...
& $taskGo build -ldflags=-H=windowsgui -o dist/prototype/MofuMouse-Prototype-Flow225.exe ./cmd/petprototype
# SmokeはStart-Process -WindowStyle Hiddenで -size 64/96 -smoke 7s -report 絶対パスを渡す。
# 終了キー登録が競合するため旧試作は停止してから実行する。
```

終了はCtrl+Alt+Shift+Q、または通知領域「試作版を終了」。コマ数/動物選択は現在degu用の範囲があるため、新種を無条件にprepare helperへ流さない。

## 使っている補助スクリプトと境界

| Script | 役割 / 限界 |
|---|---|
| `build_flow_walk.py` | デグー動画1.5〜2.75秒のdecode/IS-Net/30枚出力v1。`--pilot`のみ追加引数 |
| `build_animal_cursor_pilot.py` | 旧独立ImageGen素材のpacker。GIF保存・再読込・時間表helperは再利用。動画にImageGen原本契約を偽装しない |
| `prepare_coat_sheet_pilot.py` | 既存8ポーズを固定2×4へ配置。source pixelsに新しい脚や色を描かない |
| `export_coat_sheet_pilot.py` | 最初のブルー生成1枚を切り出し・共通逆変換・初回833ms出力。生成原本パス固定 |
| `retime_flow_walk.py` | 30枚を保ち1.5/2.25倍へ。短delay時のGIFのみ互換間引き |
| `retime_coat_sheet_pilot.py` | ブルー8枚を再生成せず556msへ。旧/新APNGのRGBA一致確認 |
| `prepare_pet_prototype.py` | 検証済みPNGをprototypeへコピーしactiveを書換。読取専用コマンドではない |
| `run_pet_prototype.ps1` | 同じパスで既に起動中なら重複起動しない。既定96px |
| `check_animal_cursor_state.py` | 現在選択と指定coatのhash・時間表・アニメ読込をread-onlyで監査 |

現在のWindows環境ではPythonは `C:/Users/User/.cache/codex-runtimes/codex-primary-runtime/dependencies/python/python.exe`、Goは `C:/Program Files/Go/bin/go.exe`。実在を確認してtask-specific変数に入れる。Skill構文検証だけは `C:/Users/User/.codex/venvs/skill-tools/Scripts/python.exe` を使う。全て機械固有の観測値であり別PCでは再確認する。

ローカル動画/alpha依存は `.codex/tools/animal-cursor-video/` と `.codex/tools/animal-cursor-matting/`。IS-Netのこの実績ではMD5 `fc16ebd8b0c10d971d3513d564d01e29` を検証。モデルを差し替えた場合は出典・hash・結果を新しく記録する。

## 出典記録と再開地点

### 10種類制作で追加した共通手順（2026-09-28）

基準の正本台帳は `assets/manifest/animal-cursor-ten-species.json`。完成条件は `docs/exec-plans/active/animal-cursor-ten-species.md`。Flow方式の基準動画と、個別ImageGen方式の色違いポーズを別の原本種別として記録する。

1. 新種のpromptを `assets/source/animal-cursor-v2/<species>/<coat>/identity-flow-v1.prompt.txt` に保存し、内蔵ImageGenを1枚生成。`register_animal_identity.py --species ... --coat ... --original <絶対PNG>` で同hashコピーと由来を保存。
2. Flowへ入力をアップロードし、開始/終了画像、4秒、720p、x1を画面で確認。promptを `<species>/<coat>/flow/<action>-attempt-01.prompt.txt` に保存。既存の候補へ上書きせずattemptを増やす。
3. 完成動画はブラウザのdownloadMediaで保存。`register_flow_job.py --species ... --coat ... --action walk|idle --attempt 1 --download <絶対MP4> --url <観測した編集URL>` で原動画と設定を登録。
4. `build_flow_animal.py --config <JSON> --inspect` で元コマ一覧。歩行は足と尾の位相に基づきstart/countを設定し、通常24fpsを保持。待機は原則96コマ/4秒。登録時のselection_statusは依頼時の状態で、最終状態は基準台帳/preview manifestに記録する。
5. 同コマンドの `--pilot` で先頭/中央/末尾を確認してから全出力。歩行周期の変更は新revisionを使う。既存decode/sourceが異なる動画・時刻・hashなら失敗させる。原RGBを保持し、全コマ共通crop/scale/offsetでalphaのみを処理する。
6. 原寸/64/96px、白/暗、全シーケンスと継ぎ目を目視し、`VISUAL_REVIEW.md` を書く。`register_reviewed_motion.py --manifest <preview/manifest.json> --notes <VISUAL_REVIEW.md>` で再読込・hash検証後、基準台帳へready登録。生成・ビルド成功だけではreadyにしない。
7. `prepare_animal_coat_handoff.py --manifest ... --species ... --motion ... [--indices <8番号>]` で比較用8ポーズとdurationを保存。待機の閉眼を平均間隔で落とさない。`publish_animal_coat_handoff.py --species ... --motion ... --coats <2候補>` で親handoffを公開。
8. 色違いは元frameを1画像1poseで直接編集するのが既定。最新ユーザー指示で明示許可された4pose例外は段階5Aのbounded試験として扱う。30実ポーズの歩行基準、旧8セル履歴、4コマ試験の状態と実生成回数を混同しない。
9. `render_animal_cursor_catalog.py` で一覧、`audit_animal_cursor_catalog.py` でready件の原動画/source/prompt/manifest/48等の実exportを照合。GIFの表示コマ数はファイル自身から読む（デグーはAPNG30とGIF28が異なる）。

作業中断・長大化時は `.codex/tasks/animal-cursor-ten-species-resume.md` を更新し、Goalを完了扱いにしない。色違い担当所有の台帳・スクリプトへ親が書き込まない。

各生成/変換で次を保存する。全候補が同一schemaであると仮定せず、実scriptとmanifestを読む。

- 対象種/色/動作、generation tool、prompt、参照と役割、原本と選択コピーのhash。
- 動画なら元fps、採用時刻、元frame番号、alpha方法、RGB保持の検証。
- シートなら生成1枚のhash、行列/セル境界、セル→元ポーズ、共通変換。
- PNG/APNG/GIFのサイズ、実コマ数、delay表、loop、alpha、出力hash。
- 原寸/64/96px/白暗/loopの目視結果、未確認点、候補か採用かruntime反映済みか。
- Windows反映時は使用manifest、exehash、実行ログと実PIDを観測時点付きで記録。

再開時に読む小さい台帳へ「完了物・担当・次の1作業・未確認点」を書く。原本、旧速度、退避exeを残す。未検証の計画を完了済み一覧に混ぜない。

## 過去の失敗から残す判断

- 独立再生成した16コマはユーザーが不自然と評価。タイミングだけでなく元動画と別のポーズ列になったことを踏まえ、現在は元動画を直接使う。
- 初期8コマの少なさへの不満と、後のブルー8コマ試作への好評価は別の評価対象。全動物に8枚を強制しない。
- 透過画像をFlowへ渡して出た暗いもやは、灰背景の別入力で改善した。Flow出力は依然として不透明で、透過処理は別工程。
- 8セルの色替えは形状IoUが高くても、顔や毛並みまで完全同一ではなかった。定量スコアだけで全色を採用しない。
- Windows smokeのキーは `exit_hotkey_registered`。`hotkey_registered`という不存在キーを読んで誤失敗した履歴がある。
- Codex read_threadが失敗しても、list_threadsとsend_messageが動く場合があった。送信成功後はwait_threadsでactive/completeを確認し、読めないからと同じ生成依頼を重複送信しない。
