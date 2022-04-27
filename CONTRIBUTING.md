# 開発に関して

## 前提

controllerの作成にkubebuilderを使用しています。  
https://github.com/kubernetes-sigs/kubebuilder  
よって、基本的な開発手順はkubebuilderのドキュメントにある通りになります。  
よく使うコマンドやひと手間入れる部分などを書いておきます。

## 開発

### 起動

`make run` を実行するとcontrollerが起動し、kubeconfigのデフォルトcontextに設定されているクラスタの`MetricsSource` resourceのwatchを開始します。  
動作確認をする場合は、そのクラスタに `MetricsSource` をapplyして、statusを確認します。  
メトリクスはcontrollerが出力するので、クラスタの場所に関係なくlocalhostにアクセスしてください。

### CRD変更

CRDは、 `api/v1/metricssource_types.go` を変更してコマンド実行 `make crd` することで `manifest/deploy/crd.yaml` が更新されます。  
同時に `config/crd/bases/k8s.oder.com_metricssources.yaml` も更新されますが、無視してください。  
本来、 `config` ディレクトリはこのようにkubebuilderによって自動生成されるmanifestが入るのですが、今回の用途では不要なものも大量に作られてしまうので、必要なものを `manifest/` に移動した結果こうなっています。  

### テスト

`make test` で実行できます。  
やることは go test なのでそちらでも構いません。  
`controllers/suite_test.go` はkubebuilderが自動生成するもので、内容は把握してないです。

### イメージのビルド

`make docker-build` でimageをbuildします。  
環境変数 `IMG` にリポジトリとタグを指定します。  
結局のところ docker build しているだけなので直接そちらでよいです。

同様に、 `make docker-push` もありますが、やはり docker push なのでどちらでも。
