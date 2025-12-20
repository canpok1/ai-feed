---
name: get-pr-review-thread-details
description: 指定したレビュースレッドIDの詳細情報を取得し、人間が読みやすい形式で表示します。スレッド内の全コメントを時系列順で確認でき、コンテキストの把握に役立ちます。
---

## 使用方法

### スクリプトの実行

```bash
./.claude/skills/get-pr-review-thread-details/get-pr-review-thread-details.sh <スレッドID> [<スレッドID2> ...]
```

### 引数

- `<スレッドID>`: 確認したいレビュースレッドのID（例: `PRRT_kwDONTZR484BhKaH`）
- 複数のスレッドIDを指定することも可能（スペース区切り）

### 出力形式

各スレッドについて以下の情報を整形して表示します：

```
═══════════════════════════════════════════════════════════════════════════════
スレッドID: PRRT_kwDONTZR484BhKaH
解決状態: ✗ 未解決 / ✓ 解決済み
ファイル: internal/app/recommend.go
行番号: 42-45
───────────────────────────────────────────────────────────────────────────────
コメント一覧（時系列順）:

[1] reviewer-username - 2025-12-10T10:30:00Z

    コメント本文

[2] developer-username - 2025-12-10T11:00:00Z

    返信コメント本文
═══════════════════════════════════════════════════════════════════════════════
```

### 使用例

```bash
# 単一スレッドの詳細を取得
./.claude/skills/get-pr-review-thread-details/get-pr-review-thread-details.sh "PRRT_kwDONTZR484BhKaH"

# 複数スレッドの詳細を同時に取得
./.claude/skills/get-pr-review-thread-details/get-pr-review-thread-details.sh "PRRT_kwDONTZR484BhKaH" "PRRT_kwDONTZR484BhKaI"
```

### 次のアクション

スレッドの内容を確認した後：
- 対応が必要な場合: コードを修正してコミット・プッシュ
- 返信する場合: `reply-to-review-thread` スキル
- 解決する場合: `resolve-pr-thread` スキル
