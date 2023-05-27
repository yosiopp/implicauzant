# Implicauzant
OIDCのImplicit FlowのみをサポートするIdPサーバとして動作し、匿名ユーザに対してIDトークンを発行します。

このアプリケーションは、OIDC Implicit Flowを用いた要認証アプリケーションに対して匿名ログインを可能とすることを目的とした実験的実装です。  
セキュリティ的な問題を含んでいる可能性があります。  
また、継続的な開発や保守は予定していません。

## Request
```
GET /authorize?
    scope=openid+profile                    // openid または openid+profile
    &response_type=id_token                 // id_token 固定
    &client_id=https://api.example.com      // 利用者を示すRPのURI
    &redirect_uri=https://app.example.com/  // RPのリダイレクト先URI
    &state=c3DNysWCpz6L                     // レスポンスを紐づけるセッションID
    &nonce=kFMh3wWRDrTr
```

Implicauzantは要求されたリクエストに基づき、認証画面を表示します。  
認証画面にてnameとsecretを入力してください。  
nameとsecretの組み合わせを元にハッシュ計算してsubクレームを生成します。  
redirect_urlに指定されたURIにリダイレクトします。  
認証画面にて同じnameとsecretの組み合わせを入力した場合、常に一意のsubクレームが返されます。  
※ユーザを認証しているわけでは無い点に注意してください。

## Response
```
HTTP/1.1 302 Found
Location: https://app.example.com/#
    token_type=Bearer
    &id_token=eyJ...
    &state=c3DNysWCpz6L
```

## ID Token structure
IDトークンのペイロードには以下の情報が含まれます。
```
{
    "sub": "???????????",               // nameとsecretを元にしたhash値
    "iss": "Implicauzant",              // Implicauzant 固定（※環境変数で変更可能）
    "aud": "https://api.example.com",   // requestで渡されたclient_idの値 
    "exp": 1482809609,                  // 発行日時+86400（※環境変数で変更可能）
    "iat": 1482773609,                  // 発行日時
    "nonce": "kFMh3wWRDrTr",            // requestで渡されたnonceの値
    "name": "name"                      // 入力したnameの値（scopeにprofileを付与した場合のみ）
}
```

なお、生成されるJWTの署名アルゴリズムは HS256 固定です。

## RP側の対応
IdPに対して事前のクライアント登録などの設定は不要です。  
RPは必要なクエリストリングを指定してimplicauzantのエンドポイントにリクエストを送信してください。

ImplicauzantからのIDトークンを受け入れる場合、RPは以下の検証を実施してください。
* JWTの署名が正当であること
* expクレームが未来日時であること
* issクレームが"Implicauzant"であること
* audクレームがRequest時に設定した値であること
* nonceクレームがRequest時に設定した値であること

## 環境変数
|環境変数|説明|デフォルト値|
|:--|:--|:--|
|IMPLICAUZANT_SECRET_KEY| (*必須*) 署名に利用するシークレット|-|
|IMPLICAUZANT_SALT| (*必須*) subクレームを生成する際に利用するSALT値|-|
|IMPLICAUZANT_SUBJECT_POSTFIX|subクレームの接尾子|@implicauzant|
|IMPLICAUZANT_ISSUER|issに設定される値|Implicauzant|
|IMPLICAUZANT_EXPIRES_IN|expに使用される有効期限（秒）|86400|


## その他
命名は「暗黙ユーザ」を意味するエスペラント語 __Implica uzanto__ から。