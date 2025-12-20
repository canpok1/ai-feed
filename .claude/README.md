# .claude ディレクトリ構成ガイド

このディレクトリには、Claude Codeの動作を制御する設定ファイルやリソースが含まれています。

## ディレクトリ構成

```
.claude/
├── agents/           # 専門家エージェント定義
├── commands/         # スラッシュコマンド定義（ユーザーが起動）
├── rules/            # ファイルタイプ別コーディング規約
├── skills/           # 再利用可能なタスク/スクリプト
├── CLAUDE.md         # Claude Code向けガイダンス
└── settings.json     # 権限・フック設定
```

## 各ディレクトリの役割

### agents/（専門家エージェント）

**役割**: プロジェクト固有のルールに基づいたレビュー・相談・ガイダンスを提供

**使用タイミング**:
- 実装前: 設計や方針についてガイダンスを依頼
- 実装中: ルールの解釈や適用方法について相談
- 実装後: コードのレビューを依頼

**定義されているエージェント**:
- `coding-specialist`: コーディングルール（docs/01_coding_rules.md）に関する専門家
- `architecture-specialist`: アーキテクチャルール（docs/02_architecture_rules.md）に関する専門家
- `testing-specialist`: テストルール（docs/03_testing_rules.md）に関する専門家
- `document-specialist`: 文書化ルール（docs/07_document_rules.md）に関する専門家

### rules/（ファイルタイプ別規約）

**役割**: ファイル編集時に自動的に適用されるコーディング規約

**適用方法**: YAMLフロントマターの`paths`で対象ファイルを指定

**定義されているルール**:
- `go-coding.md`: `**/*.go` に適用されるGo言語のコーディング規約
- `go-testing.md`: `**/*_test.go` に適用されるテスト規約
- `documentation.md`: `**/*.md` に適用されるドキュメント作成規約

### skills/（再利用可能なタスク）

**役割**: 特定の作業を自動化するスクリプトとガイドライン

**構成**:
- SKILL.md（ドキュメント）: スキルの説明と使用方法
- スクリプト（オプション）: 実際の処理を行うシェルスクリプト

**定義されているスキル**:
- `create-pr`: プルリクエスト作成の自動化
- `get-pr-review-comments`: PRの未解決レビューコメント取得
- `get-pr-review-thread-details`: レビュースレッドの詳細情報取得
- `reply-to-review-thread`: レビュースレッドへの返信投稿
- `resolve-pr-thread`: レビュースレッドの解決
- `plan-issue`: GitHub issue対応タスクの作成支援
- `plan-pr`: PRレビューコメント対応タスクの作成支援
- `review`: 実装内容のレビュー実施ガイドライン

### commands/（スラッシュコマンド）

**役割**: 複数のスキル・エージェントを組み合わせた定型ワークフロー

**使用方法**: **ユーザーが手動で起動** (`/command-name [引数]`)

**特徴**: 決まった手順を段階的に実行し、複雑な作業フローを簡素化

**定義されているコマンド**:
- `/handle-issue [issue番号]`: issue対応の全体フロー（計画→実行→レビュー→PR作成）
- `/handle-pr-comment [PR番号]`: PRコメント対応フロー（コメント確認→計画→実行）
- `/refine-issue [issue番号]`: GitHub issueの内容を精緻化
- `/run`: tmp/todo内のタスクを実行

## 使い分けのガイドライン

| 用途 | 使用するもの | 読み込みタイミング |
|------|------------|-------------------|
| コーディング規約の自動適用 | `rules/` | ファイル編集時に自動読み込み |
| レビュー・相談 | `agents/` | Claude Codeまたはユーザーが必要時に使用 |
| 単一タスクの自動化 | `skills/` | Claude Codeまたはユーザーが必要時に使用 |
| 定型ワークフロー実行 | `commands/` | **ユーザーが手動実行** |

## 典型的な使用例

### 新機能実装時

1. `/handle-issue #123` でissue対応フローを開始（ユーザーが手動実行）
2. Claude Codeが `plan-issue` スキルを使用してタスク作成
3. `/run` でタスクを実行
4. 実装時に `rules/` が自動適用されてコーディング規約を遵守
5. `coding-specialist` などのエージェントでレビュー
6. `create-pr` スキルでPR作成

### PRレビュー対応時

1. `/handle-pr-comment #456` でPRコメント対応フローを開始（ユーザーが手動実行）
2. Claude Codeが `get-pr-review-comments` スキルでコメント取得
3. `plan-pr` スキルで対応タスク作成
4. `/run` でタスクを実行
5. `reply-to-review-thread` や `resolve-pr-thread` スキルで対応完了

## 関連ドキュメント

- [CLAUDE.md](./CLAUDE.md): Claude Code向けの全体的なガイダンス
- [settings.json](./settings.json): 権限設定とフック設定
- [docs/](../docs/): プロジェクト全体のドキュメント
