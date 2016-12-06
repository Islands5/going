
### 動作確認環境
MacOS Sierra
Go1.7.1(1.7未満だと動かないかも...)

### インストール

```
$go get github.com/islands5/going
```

これでgoingコマンドが使えるようになります。

### データベースの準備

mysqlがPCに入っている方は、ターゲットのデータベースが作成されているか確認して次のステップへお進みください
mysqlがPCに入ってない方は、docker-composeを使って立ち上げます。

[Docker for Mac](https://docs.docker.com/docker-for-mac/)
こちらをインストールしていただいて、

```
$cp -r $GOPATH/src/github.com/islands5/going/sample ./
$cd sample
$docker-compose up

#別のターミナルを立ち上げる
$docker exec -it sample_db /bin/bash
#=>dockerコンテナの中へ
$mysql -u root -ppassword
#CREATE DATABASE going_demo
```

これでテスト環境ができました

### 初期化
どこでもいいので作業ディレクトリへ移動して

```
$going init
```

を実行してください
カレントディレクトリにgoing-assetsというディレクトリが作成されます。

### 設定&SQLの記述、反映

going-assets/going.ymlにデータベースの設定を記述します。
            /sql内に実行したい*.sqlファイルを命名規則に従って配置していきます。
規則は
V{バージョン名}__{適当な名前}.sql
サンプルのファイルは

```
V1__create_person_table.sql
```

という風になっています

コマンドを実行してみます。

```
$going up
#=> applying: V1__create_person_table.sql...
```

これでDBに反映されます。

続いて、Siteテーブルを追加します。
例によってgoing-assets/sql/V2__create_site_table.sqlを作成します。

```
$going up
#=> already applied: V1__create_person_table.sql
applying: V2__create_site_table.sql...


データベース
+----------------------+
| Tables_in_going_demo |
+----------------------+
| person               |
| site                 |
+----------------------+
2 rows in set (0.01 sec)
```

### 削除!

```
$going reset
```

!!使うときはデータベースが跡形もなく消えるので気をつけてください

inspired php製シンプルマイグレーションツール(https://github.com/brtriver/dbup)
