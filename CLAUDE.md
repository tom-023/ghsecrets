# CLAUDE.md
このファイルは、このリポジトリのコードを扱う際のClaude Code (claude.ai/code)向けのガイダンスを提供します。

## 開発コマンド

### ビルドと実行
```bash
make build              # ghsecretsバイナリをビルド
make install            # ビルドして$GOPATH/binにインストール
make run                # アプリケーションをビルドして実行
make clean              # ビルド成果物とカバレッジファイルを削除
```

### テスト
```bash
make test               # -shortフラグ付きでユニットテストを実行
make test-integration   # 統合テストを含むすべてのテストを実行（AWS/GCP認証情報が必要）
make test-coverage      # カバレッジレポートを生成（coverage.htmlを作成）

# 特定のテストを実行
go test -v ./cmd/ghsecrets -run TestRestoreAWS
```

### コード品質
```bash
make fmt                # gofmtを使用してコードをフォーマット
make lint               # golangci-lintを実行（要インストール）
make deps               # 依存関係のダウンロードと整理
```

## アーキテクチャ概要
このCLIツールは、自動クラウドバックアップ機能を備えたGitHub Secretsを管理します。アーキテクチャは以下の主要な原則に従います：

1. コマンド構造:
コマンド（push、restore、list）は、Cobraを使用してcmd/ghsecrets/に実装されています。各コマンドファイルには独自のフラグと実行ロジックが含まれています。

2. サービス統合:
- AWS Secrets Managerは、複数のシークレットを単一のAWSシークレット内にJSONとして保存します
- 設定はViperを通じて管理され、設定ファイル、環境変数、CLIフラグをサポートします

3. 認証フロー:
- GitHub: 設定ファイルのトークン → 環境変数（GITHUB_TOKEN） → GitHub CLI（gh auth）
- AWS: 標準認証チェーン（環境変数 → credentialsファイル → SSO → IAMロール）
- 設定の優先順位: CLIフラグ → 環境変数 → 設定ファイル

4. セキュリティ機能:
- インタラクティブプロンプトはgolang.org/x/termを使用して機密入力を隠します
- すべてのGitHub secretsは送信前に暗号化されます
- AWS JSONクライアントは操作前にシークレットの存在を要求します


## 主要な実装詳細

- AWS JSON保存: json_client.goは基本のAWSクライアントをラップし、複数のキー・バリューペアを単一のシークレット内にJSONとして保存します。これにはAWSシークレットが事前に存在する必要があります。

- エラー処理: 常にコンテキストを含めてエラーをラップします。AWS JSONクライアントは操作前にシークレットの存在とJSON形式を検証します。

- テスト: AWSとGitHubの両方にモッククライアントが提供されています。ユニットテストはこれらのモックを使用し、統合テストには実際の認証情報が必要です。

- インタラクティブモード: -kまたは-vフラグが省略された場合、ツールは入力を求めます。キーは表示され、値は隠されます。

## 設定
ツールはデフォルト設定にghsecrets.yamlを使用します。
主要なセクション：
- github: token、owner、repo
- aws: region、profile、secret_name
- gcp: project、credentials_path

AWSのsecret_nameは重要です - これはすべてのGitHub secretsをJSONとして保存する単一のAWSシークレットを指定します。
