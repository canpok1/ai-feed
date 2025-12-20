---
name: resolve-pr-thread
description: 指定したレビュースレッドを「解決済み（resolved）」状態に変更します。レビュワーの指摘に対応し、承認を得た後にスレッドをクローズする際に使用します。
---

## 使用方法

### スクリプトの実行

```bash
./.claude/skills/resolve-pr-thread/resolve-pr-thread.sh <スレッドID>
```

### 引数

- `<スレッドID>`: 解決したいレビュースレッドのID（例: `PRRT_kwDONTZR484BhKaH`）
  - スレッドIDは引用符で囲んで指定することを推奨

### 出力例

```bash
レビュースレッドをresolve中...
スレッドID: PRRT_kwDONTZR484BhKaH

✓ レビュースレッドをresolveしました。
スレッドID: PRRT_kwDONTZR484BhKaH
```

### 使用例

```bash
# スレッドを解決済みにする
./.claude/skills/resolve-pr-thread/resolve-pr-thread.sh "PRRT_kwDONTZR484BhKaH"
```

### 注意事項

- スレッドをresolveする権限が必要です（通常、PR作成者またはレビュワー）
- 既にresolve済みのスレッドに対しても実行可能です
- レビュワーの承認を表すコメントに対してのみ使用してください
- 対応が必要な指摘に対しては、まず修正とコミット・プッシュを行ってからresolveしてください

### 典型的なワークフロー

1. レビューコメントを確認: `get-pr-review-comments` スキル
2. スレッドの詳細を確認: `get-pr-review-thread-details` スキル（必要に応じて）
3. 指摘事項を修正してコミット・プッシュ
4. 修正内容を返信: `reply-to-review-thread` スキル（オプション）
5. スレッドを解決: このスキル
