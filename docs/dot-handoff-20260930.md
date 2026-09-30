# dot への MofuMouse 引き継ぎ — 2026-09-30

## 入口と停止状態

ユーザーは現在の制作を停止し、dot へ引き継ぐよう指示した。監督チャットとデグー制作の A/B/C 三担当は全て作業を保存して終了した。3 分間隔の自動監督（automation 10）も削除済み。旧担当を自動再開せず、旧 stage の未消費枠を新しい生成許可として扱わない。

この GitHub ブランチは停止時点の**公開用引き継ぎ**である。作業中のローカル checkout `D:/開発/MofuMouse` は変更を保持したまま。GitHub の `main` は Electron プレビューを含む `ea2eaf3`、ローカル `main` はその 1 commit 前で大量の未コミット素材がある。リセット、上書き、公開版の更新は行っていない。

## 現状

- 対象はデグーの 10 外見。既存アグーチ以外の全 30 実ポーズ歩行と同色 idle、source provenance、ループ、runtime、Windows 実確認は未完成。`42 外見`という旧全種の範囲をデグー 10 外見の完成数と混同しない。
- A（blue）: ImageGen 実累計 44。直近 A13 は生成 0・予約 0 で停止。11 の別方式 PREP は保存済みだが入力の目視 QA と 02/03/04/06/07 再利用確認が未完了。05、11、21 の既存 HOLD と旧 4 ポーズ方式の HOLD を維持。新 10 は個別候補 QA 済みだが親 source/runtime 採用前。
- B（sand/white）: white walk の ImageGen 実累計 34。直近 B13 で 28–30 を各 1 回生成し、原本・実 prompt・receipt・個別 QA を保存。30 枚対応表と QA 用 APNG（30 フレーム、556 ms）は作成したが、全ループ、30→01、歩行↔idle の視覚受入、idle 制作、runtime 採用は未実施。sand の既存 HOLD を保持。
- C（pied）: pied lane の ImageGen 実累計 16。black_pied walk12 の修正候補は個別/01–12 比較で PASS/watch、親 source/runtime 採用前。13–15 は PREP ファイルまで保存、手入力 QA 未完、生成・予約 0。他の白斑外見を完成扱いしない。
- アプリ: GitHub の公開プレビュー `v0.1.0-preview.1` はローカルの後続素材より古い。10 種・19 外見の既存 Electron catalog と、新しいデグー候補素材を区別する。追加の Actions、リリース、Pages 公開、Mac 実機確認は行っていない。

## ローカルの正本と dot の再開点

同じ checkout を使える場合、最初に次を読む。`.codex/tasks` と `.codex/qa` は Git の ignore 対象で、この公開ブランチには含めていない。新しい clone だけでは原本と QA 記録を再現できない。

1. `.codex/tasks/degu-sol61-coordination-20260930.md` と `.codex/tasks/degu-sol61-dispatch-20260930.json`。末尾の**ユーザー停止**が過去の継続指示より新しい。
2. `.codex/tasks/degu-sol61-a-progress-20260930.md`、`degu-sol61-b-progress-20260930.md`、`degu-sol61-c-progress-20260930.md` と `.codex/qa/degu-sol61-parent-checkpoint-20260930-0842/REPORT.json`。各担当の STOP/FINAL 証跡を辿る。
3. `docs/animal-cursor-production-priorities.md`、`docs/workflows/animal-cursor-production.md`、`AGENTS.md`、`.agents/skills/mofumouse-animal-animation/SKILL.md`。最新の停止指示とユーザーからの直接指示を優先する。
4. `assets/manifest/animal-cursor-walk30-handoffs.json`、`assets/manifest/animal-cursor-walk30-degu-colors.json` と各 source-lock / CALL receipt。候補、親 source 採用、runtime 採用を分けて照合する。

dot が制作を再開する際は、残っている旧 cap を使わず、対象の色・motion・pose ID・入力 hash・出力先・新しい上限を明示した stage を新たに記録する。現物と receipt の照合、64/96px の白暗、脚数・奥の後脚・毛色・最後→先頭の確認後に source を採用する。全 30 枚と同色 idle の受入前に runtime 完成や公開版更新を宣言しない。

## GitHub に含めない手元のデータ

ローカルには未追跡の source/preview 画像・動画、既存試作、QA 画像、生成原本のコピーが数 GB ある。特に `assets/source/walk-cycle/degu/` と `.codex/qa/degu-sol61*` は停止時点の候補と証拠を含む。このブランチはそれらの一括公開や削除を行わない。dot が別 checkout で作業する場合、原本・hash・receipt の移管を先に完了し、欠けた素材を再生成で埋めない。
